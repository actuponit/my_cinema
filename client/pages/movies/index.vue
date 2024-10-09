<template>
  <div class="min-h-screen bg-black text-white p-8 flex-1">
    <h1 class="text-3xl font-bold mb-8">Movie Collection</h1>

    <!-- Search and Filter Section -->
    <div class="mb-8 space-y-4">
      <input :value="searchParams.get('search') || ''" type="text" placeholder="Search by titles, directors..."
        class="w-full p-2 bg-gray-800 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary text-lg"
        @input="(e) => setSearchParams('search', (e.target as HTMLInputElement).value)" />

      <div class="flex flex-wrap gap-4 items-center">
        <div class="relative">
          <select :value="searchParams.get('year') || ''"
            @change="(e) => setSearchParams('year', (e.target as HTMLInputElement).value)"
            class="appearance-none p-2 pr-8 bg-gray-800 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary text-lg">
            <option value="">All Years</option>
            <option v-for="year in years" :key="year" :value="year">{{ year }}</option>
          </select>
          <div class="pointer-events-none absolute inset-y-0 right-0 flex items-center px-2 text-gray-400">
            <svg class="fill-current h-4 w-4" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20">
              <path d="M9.293 12.95l.707.707L15.657 8l-1.414-1.414L10 10.828 5.757 6.586 4.343 8z" />
            </svg>
          </div>
        </div>

        <div class="relative">
          <select :value="searchParams.get('rating') || ''"
            @change="(e) => setSearchParams('rating', (e.target as HTMLInputElement).value)"
            class="appearance-none p-2 pr-8 bg-gray-800 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary text-lg">
            <option value="">All Ratings</option>
            <option v-for="rating in ratings" :key="rating" :value="rating">{{ rating }}+ Stars</option>
          </select>
          <div class="pointer-events-none absolute inset-y-0 right-0 flex items-center px-2 text-gray-400">
            <svg class="fill-current h-4 w-4" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20">
              <path d="M9.293 12.95l.707.707L15.657 8l-1.414-1.414L10 10.828 5.757 6.586 4.343 8z" />
            </svg>
          </div>
        </div>

        <!-- Actors Filter -->
        <div class="w-72 md:ml-auto">
          <label for="actors" class="bloc text-sm">Filter by actors</label>
          <USelectMenu name="actors" id="actors" :options="acotrsOptions" value-attribute="id" :model-value="actors"
              option-attribute="name" v-model:query="arg" searchable :loading="loadingActors" :searchableLazy="true"
              @update:model-value="setSearchParams('actors', $event)" multiple
              :debounce="400" />
        </div>
      </div>

      <!-- Categories Filter -->
      <div class="space-y-2">
        <h3 class="text-xl font-semibold">Categories</h3>
        <div class="flex flex-wrap gap-2">
          <button v-for="category in genres" :key="category" @click="toggleCategory(category)" :class="[
            'px-3 py-1 rounded-full text-sm font-medium transition-colors duration-200',
            searchParams.get('categories')?.includes(category)
              ? 'bg-primary text-white'
              : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
          ]">
            {{ category }}
          </button>
        </div>
      </div>
    </div>

    <!-- Active Filters -->
    <div v-if="hasActiveFilters" class="mb-6">
      <div class="flex justify-between">
        <h3 class="text-xl font-semibold mb-2">Active Filters</h3>
        <p @click="resetFilters" class="text-primary mb-4 cursor-pointer hover:underline underline-offset-3">Clear Filters</p>
      </div>
      <div class="flex flex-wrap gap-2 items-center">
        <button v-for="filter in activeFilters" :key="filter.type + filter.value"
          @click="removeFilter({ key: filter.type, value: filter.value })"
          class="px-3 py-1 text-white rounded-full border-red-500 border-2 text-sm font-medium flex items-center space-x-1 hover:bg-opacity-80 transition-colors duration-200">
          <span>{{ filter.label }}</span>
          <XIcon class="w-4 h-4 text-red-500" />
        </button>
      </div>
    </div>

    <!-- Movies Grid -->
    <div v-if="filteredMovies.length > 0"
      class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6 mb-8">
      <MovieCard v-for="movie in movies" :key="movie.id" :movie="movie" />
    </div>
    <div v-else class="text-center text-2xl text-gray-400 py-12">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 mx-auto mb-4 text-gray-600" fill="none"
        viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
          d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
      <p>No movies found matching your criteria.</p>
      <button @click="resetFilters"
        class="mt-4 px-4 py-2 bg-primary text-white rounded-md hover:bg-opacity-80 transition-colors duration-200">
        Reset Filters
      </button>
    </div>

    <!-- Pagination -->
    <Pagination :current-page="currentPage" :total-pages="Math.ceil(totalPage / limit)" @prev-page="prevPage"
      @next-page="nextPage" />

  </div>
</template>

<script setup lang="ts">
import { XIcon } from 'lucide-vue-next';
import { genres } from '~/constants';
import { MOVIE_CASTS, MOVIES_QUERY } from '~/graphql/queries/movies';
import type { Movie } from '~/types/movie';

const { result, loading, refetch } = useQuery(MOVIES_QUERY, { where: {}, offset: 0 }, {debounce: 400})
const totalPage = computed(() => result.value?.movies_aggregate.aggregate.count || 1)

const movies = computed(() => {
  return result.value.movies.map((movie: Movie) => {
    const duration = secondToString(movie.duration || 0)
    return {
      id: movie.id,
      title: movie.title,
      releaseDate: formatDateShort(movie.published_at || ''),
      thumbnail: displayImage(movie.featured_image),
      genre: movie.genre?.toLocaleUpperCase(),
      duration: `${duration.hours}H ${duration.minutes}M`,
      rating: movie.average_rating,
    }
  });
})

const currentPage = computed(() => searchParams.value.get('page') ? parseInt(searchParams.value.get('page') as string) : 1);
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

const years = [2024, 2023, 2022, 2021, 2020, 2019, 2018, 2017, 2016, 2015];
const ratings = [4, 3, 2, 1];

const actors = computed<string[]>(()=>searchParams.value.get('actors') as string[] || []);
const awhere = { is_director: { _eq: false }, _or: [{ first_name: { _ilike: "%%" } }, { last_name: { _ilike: "%%" } }] }
const { refetch: refetchActors, loading: loadingActors, result: actorsResult } = useQuery(MOVIE_CASTS, { where: awhere }, { debounce: 400 })
const arg = ref('')
const acotrsOptions = ref([])

watch(() => arg, async () => {
  if (awhere._or && awhere._or[0] && awhere._or[0].first_name)
    awhere._or[0].first_name._ilike = `%${arg.value}%`
  if (awhere._or && awhere._or[1] && awhere._or[1].last_name)
    awhere._or[1].last_name._ilike = `%${arg.value}%`
  acotrsOptions.value = []
  await refetchActors({ where: awhere })
  if (actorsResult.value)
    acotrsOptions.value = actorsResult.value.casts.map((cast: any) => ({ id: cast.id.toString(), name: `${cast.first_name} ${cast.last_name}` }))
}, {immediate: true})


const filteredMovies = movies

const limit = 10;

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
          filters.push({ type: key, value: v, label: `${key === 'actors'?'actor_id:':''}${v}` });
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


watch(() => searchParams, async () => {
  const where = { _and: <any>[] };
  searchParams.value.forEach((value, key) => {
    if (key === 'page') return;
    if (value.length > 0) {
      if (key === 'search') {
        where._and.push({ _or: [{ title: { _ilike: `%${value}%` } }, { my_director: { first_name: { _ilike: `%${value}%` } } }, { my_director: { last_name: { _ilike: `%${value}%` } } }] })
      } else if (key === 'year') {
        let date = new Date();
        date.setFullYear(parseInt(value as string), 0, 1);
        let nextYear = new Date();
        nextYear.setFullYear(parseInt(value as string) + 1, 0, 1);
        where._and.push({ published_at: { _gte: date, _lt: nextYear } })
      } else if (key === 'rating') {
        where._and.push({ average_rating: { _gte: parseInt(value as string) } })
      } else if (key === 'categories') {
        let v = (value as string[]).map((v) => v.toLocaleLowerCase());
        where._and.push({ genre: { _in: v } })
      } else if (key === 'actors') {
        let v = (value as string[]).map((v) => ({crews: {cast_id: {_eq: parseInt(v)}}}));
        where._and.push({_and: v})
      }
    }
  });
  const offset = (currentPage.value - 1) * limit;
  await refetch({ where, offset })
}, { deep: true })
</script>