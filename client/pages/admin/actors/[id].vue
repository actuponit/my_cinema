<template>
    <div class="min-h-screen flex-1 text-gray-100 py-8 px-4 sm:px-6 lg:px-8">
      <div class="max-w-3xl mx-auto">
      <!-- Actor Information -->

      <div v-if="status === 'success'" class="bg-gray-800 shadow-md rounded-lg overflow-hidden mb-8">
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
      <USkeleton v-else-if="status === 'pending'" class="mb-4 max-w-3xl h-56" :ui="{ background: 'black-800' }" />
  
        <!-- Movies List -->
        <div v-if="status === 'success'" class="bg-gray-800 shadow-md rounded-lg overflow-hidden mb-8">
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
                <div v-if="((actor as Cast).movies as CastMovie[]).length === 0">
                  <ErrorComponent title="No Movies for this actor"/>
                </div>
                <li v-else v-for="movie, index in ((actor as Cast).crews as [{}])" :key="'crew-' + index" class="py-4">
                  <div class="flex items-center space-x-4">
                    <div class="flex-1 min-w-0">
                      <p class="text-sm font-medium text-gray-100 truncate">
                        {{ (movie as CastMovie).title }}
                      </p>
                      <p class="text-sm text-gray-400">
                        {{ formatYear((movie as CastMovie).published_at) }}
                      </p>
                    </div>
                  </div>
                </li>
              </div>
            </ul>
          </div>
        </div>
        <USkeleton v-else-if="status === 'pending'" class="mb-4 max-w-3xl h-56" :ui="{ background: 'black-800' }" />
        <!-- Add New Movie Form -->
        <div class="bg-gray-800 shadow-md rounded-lg overflow-hidden">
          <div class="p-6 text-center">
            <h2 class="text-2xl font-bold text-gray-100 mb-4 text-left">Assign New Movie</h2>
            <form @submit.prevent="addMovie">
                <USelect v-model="newMovie" option-attribute="title" value-attribute="id" searchable size="lg" class="mb-4" />
                <UButton type="submit" >
                  Add Movie
                </UButton>
            </form>
          </div>
        </div>
      </div>
    </div>
  </template>
  
<script setup lang="ts">
import { ref, reactive } from 'vue'
import { CAST_QUERY_BYID } from '~/graphql/queries/casts';
  
  const {id} = useRoute().params

  const { data, status } = useAsyncQuery<{ casts_by_pk: Cast }>(CAST_QUERY_BYID, { id })

  const actor = computed<Cast | {}>(() => data.value?.casts_by_pk || {})

  // New movie form data
  const newMovie = ref()  
  // Function to add a new movie
  const addMovie = () => {
  }
</script>