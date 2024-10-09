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
          <div v-for="movie in bookmarks" :key="movie.id" class="relative">
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
  import { DELETE_BOOKMARK_MUTATION } from '~/graphql/mutations/movie';
  import { BOOKMARKS_QUERY, MOVIE_BYID } from '~/graphql/queries/movies';
  
  const movie = {title: 'asc'}
  const { result, refetch } = useQuery(BOOKMARKS_QUERY, { movie })
  const {user} = useUser();
  const router = useRouter()
  const sortBy = ref('title')
  
  const sortOptions = [
    { label: 'Title', value: 'title' },
    { label: 'Release Date', value: 'published_at' },
    { label: 'Rating', value: 'average_rating' },
  ]
  
  const bookmarks = computed(() => {
    return result.value?.bookmarks.map((bookmark) => {
      const duration = secondToString(bookmark.movie.duration || 0)
      return {
        id: bookmark?.movie_id,
        title: bookmark?.movie?.title,
        releaseDate: bookmark?.movie?.published_at,
        genre: bookmark?.movie?.genre.toUpperCase(),
        thumbnail: displayImage(bookmark?.movie?.featured_image),
        duration: `${duration.hours}h ${duration.minutes}m`,
        rating: 0,
        totalreviews: 0,
      };
    }) || []
  })
  
  watch(() => sortBy.value, async () => {
    if (sortBy.value === 'average_rating') {
      await refetch({ movie: { [sortBy.value]: 'desc_nulls_last' } })
      return
    }
    await refetch({ movie: { [sortBy.value]: 'asc' } })
  })

  const { mutate, onDone, loading } = useMutation(DELETE_BOOKMARK_MUTATION, )
  onDone(() => {
    toast.add({
      color: 'green',
      title: 'Bookmark Removed',
      description: 'The movie has been removed from your bookmarks'
    })
  })
  const removeBookmark = async (id) => {
    await mutate({ movie_id: id, user: user.value?.id }, {
      update(cache) {
        cache.modify({
          fields: {
            bookmarks(existingBookmarks, { readField }) {
              return existingBookmarks.filter(
                (bookmarkRef) => id !== readField('movie_id', bookmarkRef)
              )
            },
          }
        })
      },
      refetchQueries: [{ query: MOVIE_BYID, variables: { id: id.toString(), user_id: user.value?.id }}]
    })
  }
  
  const goToMovies = () => {
    router.push('/movies')
  }

  definePageMeta({
    middleware: ['must-login'],
  });
</script>
  