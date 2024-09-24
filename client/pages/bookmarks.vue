<template>
    <div class="container mx-auto px-4 py-8 flex-1">
      <h1 class="text-3xl font-bold mb-8 text-gray-900 dark:text-white">My Bookmarks</h1>
  
      <div v-if="bookmarks.length > 0">
        <USelect
          v-model="sortBy"
          :options="sortOptions"
          placeholder="Sort by"
          class="mb-6 max-w-xs"
        />
  
        <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6">
          <div v-for="movie in sortedBookmarks" :key="movie.id" class="relative">
            <MovieCard :movie="movie" />
            <UButton
              color="gray"
              variant="ghost"
              icon="i-heroicons-trash"
              size="sm"
              class="absolute top-0  right-0 group-hover:opacity-100 transition-opacity duration-200"
              @click="removeBookmark(movie.id)"
            >
              Remove Bookmark
            </UButton>
          </div>
        </div>
  
      </div>
  
      <ErrorComponent
        v-else
        icon="i-heroicons-bookmark"
        title="No bookmarks yet"
        message="You haven't bookmarked any movies. Start exploring and save your favorites!"
      >
        <template #action>
          <UButton
            color="primary"
            @click="goToMovies"
          >
            Explore Movies
          </UButton>
        </template>
      </ErrorComponent>
    </div>
  </template>
  
  <script setup>
  import { ref, computed } from 'vue'
  import { useRouter } from 'vue-router'
  
  // Assume these are fetched from an API or store
  const bookmarks = ref([
        { id: 1, title: "Inception", releaseDate: new Date("2023-07-15"), genre: "Sci-Fi", thumbnail: "/placeholder.webp", duration: "2h 28m",rating: 4, totalreviews: 50 },
        { id: 2, title: "The Shawshank Redemption", releaseDate: new Date("2023-08-01"), genre: "Drama", thumbnail: "/placeholder.webp", duration: "2h 22m", rating: 5, totalreviews: 100 },
        { id: 3, title: "The Dark Knight", releaseDate: new Date("2023-08-15"), genre: "Action", thumbnail: "/placeholder.webp", duration: "2h 32m", rating: 2, totalreviews: 28 },
    ])
  
  const router = useRouter()
  const sortBy = ref('title')
  
  const sortOptions = [
    { label: 'Title', value: 'title' },
    { label: 'Release Date', value: 'releaseDate' },
    { label: 'Rating', value: 'rating' },
  ]
  
  const sortedBookmarks = computed(() => {
    return [...bookmarks.value].sort((a, b) => {
      if (sortBy.value === 'title') {
        return a.title.localeCompare(b.title)
      } else if (sortBy.value === 'releaseDate') {
        return new Date(b.releaseDate) - new Date(a.releaseDate)
      } else if (sortBy.value === 'rating') {
        return b.rating - a.rating
      }
      return 0
    })
  })
  
  const removeBookmark = (id) => {
    // In a real application, you would call an API to remove the bookmark
    bookmarks.value = bookmarks.value.filter(movie => movie.id !== id)
  }
  
  const goToMovies = () => {
    router.push('/movies')
  }
</script>
  