<template>
    <div class="min-h-screen flex-1 text-gray-100 py-8 px-4 sm:px-6 lg:px-8">
      <div class="max-w-3xl mx-auto">
        <!-- Actor Information -->
      <!-- Actor Information -->
      <div class="bg-gray-800 shadow-md rounded-lg overflow-hidden mb-8">
        <div class="p-6 sm:flex sm:items-start">
          <div class="mb-4 sm:mb-0 sm:mr-6 flex-shrink-0">
            <img :src="actor.photoUrl" :alt="actor.name" class="w-32 h-32 object-cover rounded-lg shadow-lg">
          </div>
          <div>
            <h1 class="text-3xl font-bold text-gray-100 mb-2">{{ actor.name }}</h1>
            <p class="text-gray-400 mb-4">Role: {{ actor.role }}</p>
            <p class="text-gray-300">{{ actor.bio }}</p>
          </div>
        </div>
      </div>
  
        <!-- Movies List -->
        <div class="bg-gray-800 shadow-md rounded-lg overflow-hidden mb-8">
          <div class="p-6">
            <h2 class="text-2xl font-bold text-gray-100 mb-4">Movies</h2>
            <ul class="divide-y divide-gray-700">
              <li v-for="movie in actor.movies" :key="movie.id" class="py-4">
                <div class="flex items-center space-x-4">
                  <div class="flex-1 min-w-0">
                    <p class="text-sm font-medium text-gray-100 truncate">
                      {{ movie.title }}
                    </p>
                    <p class="text-sm text-gray-400">
                      {{ movie.year }}
                    </p>
                  </div>
                </div>
              </li>
            </ul>
          </div>
        </div>
  
        <!-- Add New Movie Form -->
        <div class="bg-gray-800 shadow-md rounded-lg overflow-hidden">
          <div class="p-6 text-center">
            <h2 class="text-2xl font-bold text-gray-100 mb-4 text-left">Add New Movie</h2>
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
  
  // Sample actor data
  const actor = reactive({
    name: "Tom Hanks",
    photoUrl: "https://upload.wikimedia.org/wikipedia/commons/6/69/Tom_Hanks_2019_by_Glenn_Francis.jpg",
    role: "Director",
    bio: "Thomas Jeffrey Hanks is an American actor and filmmaker. Known for both his comedic and dramatic roles, Hanks is one of the most popular and recognizable film stars worldwide.",
    movies: [
      { id: 1, title: "Forrest Gump", year: 1994, role: "Lead" },
      { id: 2, title: "Saving Private Ryan", year: 1998, role: "Lead" },
      { id: 3, title: "Cast Away", year: 2000, role: "Lead" },
      { id: 4, title: "The Da Vinci Code", year: 2006, role: "Lead" },
      { id: 5, title: "Toy Story", year: 1995, role: "Supporting" }
    ]
  })

  // New movie form data
  const newMovie = ref()  
  // Function to add a new movie
  const addMovie = () => {
  }
  </script>