<template>
    <div class="min-h-screen bg-black text-white p-8 flex-1">
      <h1 class="text-3xl font-bold mb-8">Movie Collection</h1>
  
      <!-- Search and Filter Section -->
      <div class="mb-8 space-y-4">
        <input
					:value="searchParams.get('search') || ''"
          type="text"
          placeholder="Search movies, directors, or actors..."
          class="w-full p-2 bg-gray-800 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary text-lg"
						@input="(e) => setSearchParams('search', (e.target as HTMLInputElement).value)"
        />
  
        <div class="flex flex-wrap gap-4 items-center">
          <div class="relative">
            <select
							:value="searchParams.get('year') || ''"
              @change="(e) => setSearchParams('year', (e.target as HTMLInputElement).value)"
              class="appearance-none p-2 pr-8 bg-gray-800 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary text-lg"
            >
              <option value="">All Years</option>
              <option v-for="year in years" :key="year" :value="year">{{ year }}</option>
            </select>
            <div class="pointer-events-none absolute inset-y-0 right-0 flex items-center px-2 text-gray-400">
              <svg class="fill-current h-4 w-4" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20">
                <path d="M9.293 12.95l.707.707L15.657 8l-1.414-1.414L10 10.828 5.757 6.586 4.343 8z"/>
              </svg>
            </div>
          </div>
  
          <div class="relative">
            <select
							:value="searchParams.get('rating') || ''"
              @change="(e) => setSearchParams('rating', (e.target as HTMLInputElement).value)"
              class="appearance-none p-2 pr-8 bg-gray-800 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary text-lg"
            >
              <option value="">All Ratings</option>
              <option v-for="rating in ratings" :key="rating" :value="rating">{{ rating }}+ Stars</option>
            </select>
            <div class="pointer-events-none absolute inset-y-0 right-0 flex items-center px-2 text-gray-400">
              <svg class="fill-current h-4 w-4" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20">
                <path d="M9.293 12.95l.707.707L15.657 8l-1.414-1.414L10 10.828 5.757 6.586 4.343 8z"/>
              </svg>
            </div>
          </div>

					<!-- Actors Filter -->
					 <div class="w-72 md:ml-auto">
							<label for="actors" class="bloc text-sm">Filter by actors</label>						
						 <UInputMenu
							 v-on:update:model-value="toggleActor"
							 :options="topActors"
							 value-attribute=""
							 option-attribute=""
						 />
					 </div>
        </div>
  
        <!-- Categories Filter -->
        <div class="space-y-2">
          <h3 class="text-xl font-semibold">Categories</h3>
          <div class="flex flex-wrap gap-2">
            <button
              v-for="category in categories"
              :key="category"
              @click="toggleCategory(category)"
              :class="[
                'px-3 py-1 rounded-full text-sm font-medium transition-colors duration-200',
                searchParams.get('categories')?.includes(category)
                  ? 'bg-primary text-white'
                  : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
              ]"
            >
              {{ category }}
            </button>
          </div>
        </div>
      </div>
  
      <!-- Active Filters -->
      <div v-if="hasActiveFilters" class="mb-6">
        <h3 class="text-xl font-semibold mb-2">Active Filters</h3>
        <div class="flex flex-wrap gap-2 items-center">
          <button
            v-for="filter in activeFilters"
            :key="filter.type + filter.value"
            @click="removeFilter({key: filter.type, value: filter.value})"
            class="px-3 py-1 text-white rounded-full border-red border-2 text-sm font-medium flex items-center space-x-1 hover:bg-opacity-80 transition-colors duration-200"
          >
            <span>{{ filter.label }}</span>
            <XIcon class="w-4 h-4 text-red"/>
          </button>
        </div>
      </div>
  
      <!-- Movies Grid -->
      <div v-if="filteredMovies.length > 0" class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6 mb-8">
        <MovieCard
          v-for="movie in movies"
          :key="movie.id"
          :title="movie.title"
          :date="movie.date"
          :category="movie.category"
          :thumbnailUrl="movie.thumbnailUrl"
        />
      </div>
      <div v-else class="text-center text-2xl text-gray-400 py-12">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 mx-auto mb-4 text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <p>No movies found matching your criteria.</p>
        <button @click="resetFilters" class="mt-4 px-4 py-2 bg-primary text-white rounded-md hover:bg-opacity-80 transition-colors duration-200">
          Reset Filters
        </button>
      </div>
  
      <!-- Pagination -->
      <div class="flex justify-center items-center space-x-4">
        <button
          @click="prevPage"
          :disabled="currentPage === 1"
          class="px-3 py-1 bg-primary text-white rounded-md disabled:opacity-50 hover:bg-opacity-80 transition-colors duration-200"
        >
          Previous
        </button>
        <span class="text-lg">Page {{ currentPage }} of {{ totalPages }}</span>
        <button
          @click="nextPage"
          :disabled="currentPage === totalPages"
          class="px-3 py-1 bg-primary text-white rounded-md disabled:opacity-50 hover:bg-opacity-80 transition-colors duration-200"
        >
          Next
        </button>
      </div>
    </div>
  </template>
  
  <script setup lang="ts">
  import { XIcon } from 'lucide-vue-next';
	import { ref, computed, } from 'vue';
	
	const movies = [
		{ id: 1, year:2024, title: "Inception", date: "2023-07-15", category: "Sci-Fi", thumbnailUrl: "/placeholder.webp" },
		{ id: 2, year:2024, title: "The Shawshank Redemption", date: "2023-08-01", category: "Drama", thumbnailUrl: "/placeholder.webp" },
		{ id: 3, year:2024, title: "The Dark Knight", date: "2023-08-15", category: "Action", thumbnailUrl: "/placeholder.webp" },
		{ id: 4, year:2024, title: "The Dark Knight", date: "2023-08-15", category: "Action", thumbnailUrl: "/placeholder.webp" },
	];

  const currentPage = ref(1);
  const queryMap = new Map();

	queryMap.set('search', '');
	queryMap.set('year', '');
	queryMap.set('rating', '');
	queryMap.set('categories', []);
	queryMap.set('actors', []);
	queryMap.set('page', "1");

	const {
		searchParams, 
		removeFilter, 
		resetFilters,
		setSearchParams,
	} = useSearchParams(queryMap);
  
	const categories = ['Action', 'Adventure', 'Comedy', 'Drama', 'Fantasy', 'Horror', 'Mystery', 'Romance', 'Sci-Fi', 'Thriller'];
  const years = [2024, 2023, 2022, 2021, 2020, 2019, 2018, 2017, 2016, 2015];
  const ratings = [4, 3, 2, 1];
  const topActors = [ 'actor1', 'actor2', 'actor3', 'actor4', 'actor5' ];
  
  const filteredMovies = movies
  
  const totalPages = 1;
	
  const hasActiveFilters = computed(() => {
		for (const [key, value] of searchParams.value) {
			if (key === 'page') continue;
    	if (value.length > 0) {
      	return true;
			}
		}
  	return false;  
	});
  
  const activeFilters = computed(() => {
    const filters: { type: string; value: string; label: string }[] = [];
		searchParams.value.forEach((value, key) => {
			if (key === 'page') return;
			if (value.length > 0) {
				if (key === 'search' || key === 'year')
					filters.push({ type: key, value: (value as string), label: `${key}: ${value}` });
				else if (key === 'actors' || key === 'categories') {
					for (const v of value) {
						filters.push({ type: key, value: v, label: `${v}` });
					}
				} else
					filters.push({ type: key, value: (value as string), label: `${value}` });
			}
		});
		
		return filters;
	});

  const toggleCategory = (category: string) => {
		const categories = (searchParams.value.get("categories") as string[]);
    const index = categories.indexOf(category);
    if (index === -1) {
      categories.push(category);
    } else {
      categories.splice(index, 1);
    }
    setSearchParams("categories", categories);
  };
	
  const toggleActor = (actor: string) => {
		const actors = (searchParams.value.get("actors") as string[]);
		console.log(actors);
    const index = actors.indexOf(actor);
    if (index === -1) {
      actors.push(actor);
    } else {
      actors.splice(index, 1);
    }
    setSearchParams('actors', actors);
  };

	const prevPage = () => {
    if (currentPage.value > 1) {
			const page = parseInt((searchParams.value.get('page') as string));
      if (page)
				setSearchParams('page', page - 1 + '');
    }
  };
  
  const nextPage = () => {
    if (currentPage.value < totalPages) {
			const page = parseInt((searchParams.value.get('page') as string));
			if (page)
      	setSearchParams('page', page + 1 + '');
    }
  };
  
</script>
  