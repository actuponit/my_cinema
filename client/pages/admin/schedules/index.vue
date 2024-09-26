<template>
    <div class="min-h-screen bg-gray-950 flex-1 text-gray-100 p-8">
      <h1 class="text-3xl font-bold mb-8">Movie Schedules</h1>
  
      <!-- Filtering Section -->
      <UContainer class="bg-gray-800 p-6 rounded-lg mb-8 shadow-lg">
        <h2 class="text-xl font-semibold mb-4">Filter Schedules</h2>
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <div>
            <label for="movie" class="block text-sm font-medium text-gray-400 mb-1">Movie</label>
            <!-- <UInput id="movie" name="movie" :model-value="searchParams.get('movie') as string" @input="onInput('movie', $event)"/> -->
            <input id="movie" name="movie" :value="(searchParams.get('movie') as string) || ''" type="text" placeholder="Search by movie title" class="relative block w-full disabled:cursor-not-allowed disabled:opacity-75 focus:outline-none border-0 form-input rounded-md placeholder-gray-400 dark:placeholder-gray-500 text-sm px-3.5 py-2.5 shadow-sm bg-transparent text-gray-900 dark:text-white ring-1 ring-inset ring-primary-500 dark:ring-primary-400 focus:ring-2 focus:ring-primary-500 dark:focus:ring-primary-400" @input="onInput('movie', $event)" />
          </div>
          <div>
            <label for="date" class="block text-sm font-medium text-gray-400 mb-1">Date</label>
            <UInput id="date" name="date" type="date" size="lg" :model-value="(searchParams.get('date') as string)" color="primary" @update:model-value="setSearchParams('date', $event)" />
          </div>
          <div>
            <label for="hall" class="block text-sm font-medium text-gray-400 mb-1">Hall</label>
            <USelectMenu size="lg" color="primary" id="hall" :model-value="searchParams.get('hall')"  name="hall" :options="halls" @change="setSearchParams('hall', $event)"/>
          </div>
          <div>
            <label for="format" class="block text-sm font-medium text-gray-400 mb-1">Format</label>
            <USelectMenu size="lg" color="primary" id="format" :model-value="searchParams.get('format') as string" name="format" :options="formats" @change="setSearchParams('format', $event)"/>
          </div>
        </div>
        <div class="mt-4 flex justify-end">
          <UButton variant="link" @click="resetFilters()">
            Clear Filters
          </UButton>
        </div>
        <div v-if="hasActiveFilters" class="flex flex-wrap gap-2 items-center">
          <button
            v-for="filter in activeFilters"
            :key="filter.type + filter.value"
            @click="removeFilter({key: filter.type, value: filter.value})"
            class="px-3 py-1 text-white rounded-full border-red-500 border-2 text-sm font-medium flex items-center space-x-1 hover:bg-opacity-80 transition-colors duration-200"
          >
            <span>{{ filter.label }}</span>
            <XIcon class="w-4 h-4 text-red-500"/>
          </button>
        </div>
      </UContainer>
  
      <!-- Schedules Table -->
      <UTable class="w-full" :rows="schedules" :columns="columns">
          <template #actions-data="{ row }">
              <div class="flex space-x-2 justify-around">
                <UButton @click="viewTickets(row.id)" variant="link">
                  View Tickets
                </UButton>
                <button @click="openModal(row)" class="px-3 py-1 bg-red-600 text-white rounded-md hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2 focus:ring-offset-gray-800">
                  Delete
                </button>
              </div>
          </template>
      </UTable>
      <Pagination :current-page="currentPage"  :total-page="1" @prev-page="prevPage" @next-page="nextPage"/>
  
      <!-- Confirmation Modal -->
    </div>
  </template>
  
  <script setup lang="ts">
  import { ref, computed } from 'vue'
import ConfirmModal from '~/components/ConfirmModal.vue';
import type { Schedule } from '~/types';
import { XIcon } from 'lucide-vue-next';
  
  // Sample data (replace with actual data fetching logic)
  const schedules = ref<Schedule[]>([
    { id: 1, movie: "Inception", date: "2023-06-15", time: "12:00AM", hall: "Hall A", format: "IMAX", price: 20 },
    { id: 2, movie: "The Dark Knight", date: "2023-06-15",time: "12:00AM", hall: "Hall B", format: "2D", price: 20 },
    { id: 3, movie: "Interstellar", date: "2023-06-16", hall: "Hall C", time: "12:00AM", format: "3D", price: 20 },
    { id: 4, movie: "Inception", date: "2023-06-16", hall: "Hall A", time: "12:00AM", format: "2D", price: 20 },
    { id: 5, movie: "The Dark Knight", date: "2023-06-17", hall: "Hall B", time: "12:00AM", format: "IMAX", price: 20 },
  ])

  const columns = [
    {label: "Movie", key: "movie"}, 
    {label: "Date", key: "date"}, 
    {label:"Time", key: "time"}, 
    {label:"Hall", key: "hall"}, 
    {label:"Format", key:"format"},
    {label:"Price", key:"price"},
    {label:"Actions", key: "actions"},
  ]
  
  const currentPage = ref(1)
  const halls = ["Hall A", "Hall B", "Hall C"]
  const formats = ["2D", "3D", "IMAX"]
  
  const filtersmp = new Map();
  filtersmp.set('movie', '');
  filtersmp.set('date', '');
  filtersmp.set('hall', '');
  filtersmp.set('format', '');
  
    const {
      searchParams, 
      removeFilter, 
      resetFilters,
      setSearchParams,
      hasActiveFilters
	} = useSearchParams(filtersmp);
    const movie = computed(() => (searchParams.value.get('movie') as string) || '')
    // const search = computed(() => {
    //     console.log("search here: ",searchParams.value.get('movie') as string)
    //     return searchParams.value.get('movie') as string
    // })
    const activeFilters = computed(() => {
        const filters: { type: string; value: string; label: string }[] = [];
		searchParams.value.forEach((value, key) => {
			if (key === 'page') return;
			if (value.length > 0) {
				filters.push({ type: key, value: (value as string), label: `${key}: ${value}` });
			}
		});
		return filters;
	});
//   watch(() => searchParams, () => console.log(searchParams.value), {deep: true})
  function onInput(key:string, e: Event) {
      const target = e.target as HTMLInputElement;
      return setSearchParams(key, target.value);
  }
  const modal = useModal()
  function openModal (schedule: Schedule) {
        modal.open(ConfirmModal, {
        schedule: schedule,
        onClose () {
            modal.close()
        }
    })}
    
    const prevPage = () => {
    if (currentPage.value > 1) {
			const page = parseInt((searchParams.value.get('page') as string));
      if (page)
				setSearchParams('page', page - 1 + '');
    }
  };
  
  const nextPage = () => {
    if (currentPage.value < 1) {
			const page = parseInt((searchParams.value.get('page') as string));
			if (page)
      	setSearchParams('page', page + 1 + '');
    }
  };
//   const applyFilters = () => {
//     // This function can be used to trigger any additional actions when filters are applied
//     console.log("Filters applied:", filters.value)
//   }
  
  const viewTickets = (scheduleId: number) => {
    // This function would typically navigate to the tickets page for the selected schedule
    console.log("Viewing tickets for schedule:", scheduleId)
  }

</script>