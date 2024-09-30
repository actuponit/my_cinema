<template>
    <div class="min-h-screen flex-1 text-gray-100 py-8 px-4 sm:px-6 lg:px-8">
      <div class="max-w-3xl mx-auto">
      <!-- Actor Information -->

      <div v-if="!status" class="bg-gray-800 shadow-md rounded-lg overflow-hidden mb-8">
        <div class="p-6 sm:flex sm:items-start">
          <div class="mb-4 sm:mb-0 sm:mr-6 flex-shrink-0">
            <img :src="displayImage((actor as Cast).photo_url)" :alt="(actor as Cast).first_name" class="w-32 h-32  rounded-lg shadow-lg">
          </div>
          <div>
            <h1 class="text-3xl font-bold text-gray-100 mb-2">{{ (actor as Cast).first_name }} {{ (actor as Cast).last_name }}</h1>
            <p class="text-gray-400 mb-4">Role: {{ (actor as Cast).is_director?"Director":"Actor" }}</p>
            <p class="text-gray-300">{{ (actor as Cast).bio || "NO BIO AVAILABLE" }}</p>
          </div>
        </div>
      </div>
      <USkeleton v-else-if="status" class="mb-4 max-w-3xl h-56" :ui="{ background: 'black-800' }" />
  
        <!-- Movies List -->
        <div v-if="!status" class="bg-gray-800 shadow-md rounded-lg overflow-hidden mb-8">
          <div class="p-6">
            <h2 class="text-2xl font-bold text-gray-100 mb-4">Movies</h2>
            <ul class="divide-y divide-gray-700">
              <div v-if="(actor as Cast).is_director">
                <div v-if="((actor as Cast).movies as CastMovie[]).length === 0">
                  <ErrorComponent title="No Movies for this director"/>
                </div>
                <li v-else v-for="movie, index in ((actor as Cast).movies as [CastMovie])" :key="'director-' + index" class="py-4">
                  <div class="flex items-center space-x-4">
                    <div class="flex-1 min-w-0">
                      <p class="text-sm font-medium text-gray-100 truncate">
                        {{ movie.title }}
                      </p>
                      <p class="text-sm text-gray-400">
                        {{ formatYear(movie.published_at) }}
                      </p>
                    </div>
                  </div>
                </li>
              </div>
              <div v-else>
                <div v-if="((actor as Cast).crews as any[]).length === 0">
                  <ErrorComponent title="No Movies for this actor" message=""/>
                </div>
                <li v-else v-for="movie, index in ((actor as Cast).crews as any[])" :key="'crew-' + index" class="py-4">
                  <div class="flex items-center space-x-4">
                    <div class="flex-1 min-w-0">
                      <p class="text-sm font-medium text-gray-100 truncate">
                        {{ movie.movie.title }}
                      </p>
                      <p class="text-sm text-gray-400">
                        {{ formatYear(movie.movie.published_at) }}
                      </p>
                    </div>
                  </div>
                </li>
              </div>
            </ul>
          </div>
        </div>
        <USkeleton v-else-if="status" class="mb-4 max-w-3xl h-56" :ui="{ background: 'black-800' }" />
        <!-- Add New Movie Form -->
        <div v-if="!(actor as Cast).is_director" class="bg-gray-800 shadow-md rounded-lg">
          <div class="p-6 text-center">
            <h2 class="text-2xl font-bold text-gray-100 mb-4 text-left">Assign New Movie</h2>
            <form @submit.prevent="addMovie">
                <USelectMenu 
                  v-model="newMovie" 
                  placeholder="Search movie title" 
                  option-attribute="title" 
                  value-attribute="id"
                  :searchable="search" 
                  size="lg" 
                  class="mb-4"
                  :loading="loading"
                  :searchableLazy="true"
                  :popper="{ placement: 'top-end' }"
                 />
                <UButton :loading="addLoading" type="submit" >
                  Add Movie
                </UButton>
            </form>
          </div>
        </div>
      </div>
    </div>
  </template>
  
<script setup lang="ts">
import { ref } from 'vue'
import { CAST_QUERY_BYID, CAST_QUERY_MOVIE } from '~/graphql/queries/casts';
  
  const {id} = useRoute().params

  const { result:data, loading:status, } = useQuery<{ casts_by_pk: Cast }>(CAST_QUERY_BYID, { id }, {
    errorPolicy: 'ignore',
    notifyOnNetworkStatusChange: true,
  },)

  const actor = computed<Cast | {}>(() => data.value?.casts_by_pk || {})

  const where = {_and: {title: {_ilike: `%%`}, _not: {crews: {cast_id: {_eq: id}}}}}

  const { result, loading, refetch } = useQuery<{ movies: CastMovie[] }>(CAST_QUERY_MOVIE, { where })

  const search = async (arg: string) =>{
    await refetch({ where: {_and: {title: {_ilike: `%${arg}%`}, _not: {crews: {cast_id: {_eq: id}}}}}})
    if (result.value?.movies) {
      return result.value.movies.map(res => ({ id: res.id, title: res.title }))
    }
    return []
  }
  
  // New movie form data
  const newMovie = ref<number>(-1)  
  // Function to add a new movie
  const { executeInsert, loading: addLoading, onDone} = useAssignMovie()
  
  const toast = useToast()
  onDone(async ()=>{
    toast.add({
      color: 'green',
      title: 'Actor Assigned',
      description: 'Your actor is now part of the selected movies crew'
    })
    newMovie.value = -1
    useRouter().replace('/admin/actors/'+id)
    await refetch(where,)
  })

  const addMovie = async () => {
    console.log(newMovie.value)
    if (newMovie.value !== -1) {
      await executeInsert({ cast_id: id, movie_id: newMovie.value })
      return
    }
    toast.add({
      color: 'red',
      title: 'Error',
      description: 'Please select a movie to assign'
    })
  }
</script>