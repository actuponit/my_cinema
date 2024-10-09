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
          <input id="movie" name="movie" :value="(searchParams.get('movie') as string) || ''" type="text"
            placeholder="Search by movie title"
            class="relative block w-full disabled:cursor-not-allowed disabled:opacity-75 focus:outline-none border-0 form-input rounded-md placeholder-gray-400 dark:placeholder-gray-500 text-sm px-3.5 py-2.5 shadow-sm bg-transparent text-gray-900 dark:text-white ring-1 ring-inset ring-primary-500 dark:ring-primary-400 focus:ring-2 focus:ring-primary-500 dark:focus:ring-primary-400"
            @input="onInput('movie', $event)" />
        </div>
        <div>
          <label for="date" class="block text-sm font-medium text-gray-400 mb-1">Date</label>
          <UInput id="date" name="date" type="date" size="lg" :model-value="(searchParams.get('date') as string)"
            color="primary" @update:model-value="setSearchParams('date', $event)" />
        </div>
        <div>
          <label for="hall" class="block text-sm font-medium text-gray-400 mb-1">Hall</label>
          <USelectMenu size="lg" color="primary" id="hall" :model-value="searchParams.get('hall')" name="hall"
            :options="halls" @change="setSearchParams('hall', $event)" />
        </div>
        <div>
          <label for="format" class="block text-sm font-medium text-gray-400 mb-1">Format</label>
          <USelectMenu size="lg" color="primary" id="format" :model-value="(searchParams.get('format') as string)"
            name="format" :options="formats" @change="setSearchParams('format', $event)" />
        </div>
      </div>
      <div class="mt-4 flex justify-end">
        <UButton variant="link" @click="resetFilters()">
          Clear Filters
        </UButton>
      </div>
      <div v-if="hasActiveFilters" class="flex flex-wrap gap-2 items-center">
        <button v-for="filter in activeFilters" :key="filter.type + filter.value"
          @click="removeFilter({ key: filter.type, value: filter.value })"
          class="px-3 py-1 text-white rounded-full border-red-500 border-2 text-sm font-medium flex items-center space-x-1 hover:bg-opacity-80 transition-colors duration-200">
          <span>{{ filter.label }}</span>
          <XIcon class="w-4 h-4 text-red-500" />
        </button>
      </div>
    </UContainer>

    <!-- Schedules Table -->
    <UTable class="w-full" :rows="rschedules" :loading="loading" :columns="columns">
      <template #actions-data="{ row }">
        <div class="flex space-x-2 justify-around">
          <UButton @click="viewTickets(row.id, row.tickets)" variant="link">
            {{ row.tickets > 0 ? `View Tickets (${row.tickets})` : 'No Tickets' }}
          </UButton>
          <button @click="openModal(row)"
            class="px-3 py-1 bg-red-600 text-white rounded-md hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2 focus:ring-offset-gray-800">
            Delete
          </button>
        </div>
      </template>
    </UTable>
    <Pagination :current-page="currentPage" :total-pages="Math.ceil(totalPage / limit)" @prev-page="prevPage"
      @next-page="nextPage" />

    <!-- Confirmation Modal -->
  </div>
</template>

<script setup lang="ts">
import ConfirmModal from '~/components/ConfirmModal.vue';
import type { Schedule } from '~/types';
import { XIcon } from 'lucide-vue-next';
import { SCHEDULES_QURY } from '~/graphql/queries/schedule';
import { formats, halls } from '~/constants';
import { DELETE_SCHEDULE } from '~/graphql/mutations/schedule';
// 

const limit = 3;

const { result, loading, refetch } = useQuery<{ schedules_aggregate: { aggregate: { count: number } }, schedules: Schedule[] }>(SCHEDULES_QURY, { offset: 0, limit }, { debounce: 500 });
console.log("resul", result.value)
const rschedules = computed(() => result.value?.schedules.map((s: any) => ({
  id: s.id,
  movie: s?.movieByMovie.title,
  date: formatDateShort(s.start_time),
  time: formatTime(s.start_time),
  hall: s.hall,
  format: cinemaFormatReverse(s.format),
  price: s.price,
  tickets: s.tickets_aggregate.aggregate.count || 0
})) || [])

// const  schedules = ref<Schedule[]>(result.value?.schedules_aggregate.node || [])

const columns = [
  { label: "Movie", key: "movie" },
  { label: "Date", key: "date" },
  { label: "Time", key: "time" },
  { label: "Hall", key: "hall" },
  { label: "Format", key: "format" },
  { label: "Price", key: "price" },
  { label: "Actions", key: "actions" },
]

const filtersmp = new Map();
filtersmp.set('movie', '');
filtersmp.set('date', '');
filtersmp.set('hall', '');
filtersmp.set('format', '');
filtersmp.set('page', '1');
const {
  searchParams,
  removeFilter,
  resetFilters,
  setSearchParams,
  hasActiveFilters
} = useSearchParams(filtersmp);

const totalPage = computed(() => result.value?.schedules_aggregate.aggregate.count || 0);

const currentPage = computed(() => searchParams.value.get('page') ? parseInt(searchParams.value.get('page') as string) : 1);

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

function onInput(key: string, e: Event) {
  const target = e.target as HTMLInputElement;
  return setSearchParams(key, target.value);
}
const modal = useModal()
const toast = useToast()

const { mutate, onDone: onDeletedSchedule } = useMutation(DELETE_SCHEDULE, {
  refetchQueries: [{ query: SCHEDULES_QURY, variables: { offset: (currentPage.value - 1) * limit } }],
});
onDeletedSchedule(() => {
  toast.add({
    title: "Successful Operation",
    description: "Successfully deleted schedule!",
    color: 'green'
  })
  modal.close()
})

function openModal(sch: any) {
  modal.open(ConfirmModal, {
    title: 'Remove Schedule',
    message: `Are you sure you want to remove schedule at ${sch.start} ${sch.time} at venue ${sch.hall}?`,
    action: 'Remove',
    async onAction() {
      await mutate({ id: sch.id })
    },
    onClose() {
      modal.close()
    }
  })
}

watch(() => searchParams, async () => {
  const where = { _or: <any>[] };
  console.log(where)
  searchParams.value.forEach((value, key) => {
    if (key === 'page') return;
    if (value.length > 0) {
      if (key === 'movie') {
        where._or.push({ movieByMovie: { title: { _ilike: `%${value}%` } } })
      } else if (key === 'date') {
        const today = new Date(value as string);
        today.setHours(0, 0, 0, 0);
        const tomorrow = new Date(today);
        tomorrow.setDate(today.getDate() + 1);
        where._or.push({ start_time: { _gte: today, _lt: tomorrow } })
      } else if (key === 'format') {
        where._or.push({ [key]: { _eq: cinemaFormat(value as string) } })
      } else {
        where._or.push({ [key]: { _eq: value } })
      }
    }
  });
  const offset = (currentPage.value - 1) * limit;

  if (where._or.length > 0) {
    await refetch({ where, offset, limit });
  } else {
    await refetch({ where: {}, offset, limit });
  }
}, { deep: true })

const prevPage = async () => {
  if (currentPage.value > 1) {
    const page = parseInt((searchParams.value.get('page') as string));
    if (page) {
      setSearchParams('page', page - 1 + '');
    }
  }
};

const nextPage = async () => {
  const page = parseInt(searchParams.value.get('page') as string);
  if ((page) * limit < totalPage.value) {
    setSearchParams('page', (page + 1).toString());
  }
};

const viewTickets = (scheduleId: number, tickets: number) => {
  if (tickets > 0)
    useRouter().push(`/admin/schedules/${scheduleId}/tickets`)
  console.log("Viewing tickets for schedule:", scheduleId)
}

</script>