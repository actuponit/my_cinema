<template lang="">
    <div class="flex-1 pl-10 pr-7 min-h[100v] text-gray-100">
    <MovieCarousel />
    <FilterSection @update:filters="applyFilters"/>
    <div class="container mx-auto p-4">
      <div class="flex justify-between">
        <h1 class="mb-8 text-xl font-bold">All Schedules</h1>
        <span>{{total}} results found</span>
      </div>
      <div class="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        <USkeleton v-if="loading" v-for="(item, index) in 5" :key="index" class="h-80 w-80" />
        <MovieCard
          v-else
          v-for="movie in movies"
          type="home"
          :key="movie.id"
          :movie="movie"
        />
      </div>
    </div>
    </div>
    
    
</template>
  
<script setup lang="ts">
  import type { MovieHome } from '~/types';
import { SCHEDULES_QURY } from '../graphql/queries/schedule';

  const {result, loading, refetch} = useQuery(SCHEDULES_QURY, {offset: 0, where: {}, limit: 100});
  const total = computed(() => {result.value?.schedules_aggregate.aggregate.count})
  const movies = computed<MovieHome[]>(() => {
    console.log('res', result.value);
    return result.value.schedules.map((schedule: any) => ({id: schedule?.id, title: schedule?.movieByMovie.title, scheduleDate: formatDateShort(schedule?.start_time), price: schedule?.price, thumbnail: displayImage(schedule?.movieByMovie.featured_image), duration: formatTime(schedule?.start_time),}));
  });
  const applyFilters = async (event: any) => {
    const where = {_or: <any>[]}
    if (event.minPrice && event.minPrice >= 0 && event.maxPrice && event.maxPrice > event.minPrice)
      where._or.push({price: {_gte: event.minPrice, _lte: event.maxPrice}})
    else if (event.maxPrice && event.minPrice < event.maxPrice)
      where._or.push({price: {_lte: event.maxPrice}})
    else if (event.minPrice && event.minPrice >= 0)
      where._or.push({price: {_gte: event.minPrice}})
    if (event.selected) {
      console.log(event.selected)
      where._or.push({start_time: {
        _gte: event.selected[0],
        _lte: event.selected[1]
      }})
    }
    if (event.title && event.title !== '') {
      where._or.push({movieByMovie: {title: {_ilike: `%${event.title}%`}}})
    }
    console.log("where", where)
    if (where._or.length > 0)
      await refetch({where, offset:0, limit: 100})
    else
      await refetch({where: {}, offset: 0, limit: 100})
  } 
</script>