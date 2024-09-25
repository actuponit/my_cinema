<template>
    <div class="container mx-auto p-6">
      <h1 class="text-3xl font-bold mb-6">Add New Movie</h1>
      <Form @submit="onSubmit" :validation-schema="schema" v-slot="{ errors }">
        <div class="bg-white shadow-md rounded-lg overflow-hidden">
          <div class="border-b border-gray-200">
            <nav class="flex -mb-px" aria-label="Tabs">
              <button
                v-for="(tab, index) in tabs"
                :key="index"
                @click="activeTab = index"
                :class="[
                  activeTab === index
                    ? 'border-blue-500 text-blue-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300',
                  'w-1/4 py-4 px-1 text-center border-b-2 font-medium text-sm'
                ]"
              >
                {{ tab }}
              </button>
            </nav>
          </div>
  
          <div class="p-6">
            <div v-show="activeTab === 0">
              <div class="grid grid-cols-1 gap-6">
                <div>
                  <label for="title" class="block text-sm font-medium text-gray-700">Title</label>
                  <Field name="title" type="text" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50" />
                  <span class="text-red-500 text-xs">{{ errors.title }}</span>
                </div>
                <div>
                  <label for="releaseDate" class="block text-sm font-medium text-gray-700">Release Date</label>
                  <Field name="releaseDate" type="date" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50" />
                  <span class="text-red-500 text-xs">{{ errors.releaseDate }}</span>
                </div>
                <div>
                  <label for="genre" class="block text-sm font-medium text-gray-700">Genre</label>
                  <Field name="genre" as="select" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50">
                    <option value="">Select a genre</option>
                    <option v-for="genre in genres" :key="genre" :value="genre">{{ genre }}</option>
                  </Field>
                  <span class="text-red-500 text-xs">{{ errors.genre }}</span>
                </div>
                <div>
                  <label for="description" class="block text-sm font-medium text-gray-700">Description</label>
                  <Field name="description" as="textarea" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50" rows="4" />
                  <span class="text-red-500 text-xs">{{ errors.description }}</span>
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700">Duration</label>
                  <div class="grid grid-cols-3 gap-4">
                    <div>
                      <Field name="duration.hours" type="number" placeholder="Hours" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50" />
                    </div>
                    <div>
                      <Field name="duration.minutes" type="number" placeholder="Minutes" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50" />
                    </div>
                    <div>
                      <Field name="duration.seconds" type="number" placeholder="Seconds" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50" />
                    </div>
                  </div>
                  <span class="text-red-500 text-xs">{{ errors['duration.hours'] || errors['duration.minutes'] || errors['duration.seconds'] }}</span>
                </div>
              </div>
            </div>
  
            <div v-show="activeTab === 1">
              <div>
                <label for="director" class="block text-sm font-medium text-gray-700">Director</label>
                <Field name="director" type="text" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50" />
                <span class="text-red-500 text-xs">{{ errors.director }}</span>
              </div>
              <div class="mt-6">
                <label for="actors" class="block text-sm font-medium text-gray-700">Actors</label>
                <ActorAutocomplete v-model="selectedActors" />
              </div>
            </div>
  
            <div v-show="activeTab === 2">
              <div class="mb-4">
                <label class="block text-sm font-medium text-gray-700">Thumbnails</label>
                <div class="mt-1 flex justify-center px-6 pt-5 pb-6 border-2 border-gray-300 border-dashed rounded-md">
                  <div class="space-y-1 text-center">
                    <svg class="mx-auto h-12 w-12 text-gray-400" stroke="currentColor" fill="none" viewBox="0 0 48 48" aria-hidden="true">
                      <path d="M28 8H12a4 4 0 00-4 4v20m32-12v8m0 0v8a4 4 0 01-4 4H12a4 4 0 01-4-4v-4m32-4l-3.172-3.172a4 4 0 00-5.656 0L28 28M8 32l9.172-9.172a4 4 0 015.656 0L28 28m0 0l4 4m4-24h8m-4-4v8m-12 4h.02" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
                    </svg>
                    <div class="flex text-sm text-gray-600">
                      <label for="thumbnails" class="relative cursor-pointer bg-white rounded-md font-medium text-blue-600 hover:text-blue-500 focus-within:outline-none focus-within:ring-2 focus-within:ring-offset-2 focus-within:ring-blue-500">
                        <span>Upload files</span>
                        <Field name="thumbnails" type="file" multiple class="sr-only" />
                      </label>
                      <p class="pl-1">or drag and drop</p>
                    </div>
                    <p class="text-xs text-gray-500">PNG, JPG, GIF up to 10MB</p>
                  </div>
                </div>
                <span class="text-red-500 text-xs">{{ errors.thumbnails }}</span>
              </div>
            </div>
  
            <div v-show="activeTab === 3">
              <ScheduleForm v-model="schedules" />
            </div>
          </div>
  
          <div class="px-6 py-3 bg-gray-50 text-right">
            <button type="submit" class="inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
              Save Movie
            </button>
          </div>
        </div>
      </Form>
    </div>
  </template>
  
<script setup>
  import { ref } from 'vue'
  import { Form, Field } from 'vee-validate'
  import * as yup from 'yup'
  import ActorAutocomplete from './ActorAutocomplete.vue'
  import ScheduleForm from './ScheduleForm.vue'
  
  const tabs = ['General', 'Cast', 'Media', 'Schedule']
  const activeTab = ref(0)
  const selectedActors = ref([])
  const schedules = ref([])
  
  const genres = [
    'Action', 'Comedy', 'Drama', 'Horror', 'Romance', 'Sci-Fi', 'Thriller'
  ]
  
  const schema = yup.object().shape({
    title: yup.string().required('Title is required'),
    releaseDate: yup.date().required('Release date is required'),
    genre: yup.string().required('Genre is required'),
    description: yup.string().required('Description is required'),
    duration: yup.object().shape({
      hours: yup.number().min(0).max(10),
      minutes: yup.number().min(0).max(59),
      seconds: yup.number().min(0).max(59)
    }).test('duration', 'Duration must be greater than 0', value => 
      (value.hours || 0) * 3600 + (value.minutes || 0) * 60 + (value.seconds || 0) > 0
    ),
    director: yup.string(),
    thumbnails: yup.mixed().test('fileSize', 'File size is too large', (value) => {
      if (!value) return true
      return value.every(file => file.size <= 10 * 1024 * 1024) // 10MB
    })
  })
  
  const onSubmit = (values) => {
    console.log(values)
    // Here you would typically send the data to your backend
  }
</script>