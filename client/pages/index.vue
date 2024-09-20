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
          :title="movie.title"
          :date="movie.date"
          :category="movie.category"
          :thumbnailUrl="movie.thumbnailUrl"
        />
      </div>
    </div>
    </div>
    
    
</template>
  
<script setup>
  const movies = [
    { id: 1, title: "Inception", date: "2023-07-15", category: "Sci-Fi", thumbnailUrl: "/placeholder.webp" },
    { id: 2, title: "The Shawshank Redemption", date: "2023-08-01", category: "Drama", thumbnailUrl: "/placeholder.webp" },
    { id: 3, title: "The Dark Knight", date: "2023-08-15", category: "Action", thumbnailUrl: "/placeholder.webp" },
    { id: 3, title: "The Dark Knight", date: "2023-08-15", category: "Action", thumbnailUrl: "/placeholder.webp" },
    // Add more movies as needed
  ];
  </script>