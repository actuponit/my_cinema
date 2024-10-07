<template lang="">
    <div class="flex-1 pl-10 pr-7 min-h[100v] text-gray-100">
    <MovieCarousel />
    <FilterSection />
    <div class="container mx-auto p-4">
      <h1 class="mb-8 text-3xl font-bold">All Schedules</h1>
      <div class="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        <MovieCard
          v-for="movie in movies"
          :key="movie.id"
          :movie="movie"
        />
      </div>
    </div>
    </div>
    
    
</template>
  
<script setup lang="ts">
  import type { MovieHome } from '~/types';
  const movies: MovieHome[] = [
    { id: 1, title: "Inception", releaseDate: new Date("2023-07-15"), genre: "Sci-Fi", thumbnail: "/placeholder.webp", scheduleDate: new Date("2023-07-15"), duration: "2h 28m", price: 10, rating: 4, totalreviews: 50 },
    { id: 2, title: "The Shawshank Redemption", releaseDate: new Date("2023-08-01"), genre: "Drama", thumbnail: "/placeholder.webp", scheduleDate: new Date("2023-08-01"), duration: "2h 22m", price: 10, rating: 5, totalreviews: 100 },
    { id: 3, title: "The Dark Knight", releaseDate: new Date("2023-08-15"), genre: "Action", thumbnail: "/placeholder.webp", scheduleDate: new Date("2023-08-15"), duration: "2h 32m", price: 10, rating: 2, totalreviews: 28 },
  ];
</script>