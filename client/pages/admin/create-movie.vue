<template>
    <div class="container mx-auto p-6 flex-1">
      <h1 class="text-3xl font-bold mb-6">Add New Movie</h1>
      <UForm :state="values" @submit="onSubmit">
        <div class="p-6">
          <div class="col-span-2 bg-gray-800 p-4 mb-6 text-lg font-bold border-b border-b-gray-50">
            General info
          </div>
          <div class="grid grid-cols-2 gap-6">
            <UFormGroup label="Title" name="title" v-bind="titleProps" class="space-y-4">
              <UInput name="title" v-model="title" type="text" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50" />
            </UFormGroup>
            <UFormGroup label="Release Date" name="releaseDate" v-bind="releaseDateProps" class="space-y-4">
              <VueDatePicker name="releaseDate" v-model="releaseDate" :enable-time-picker="false" dark/>
            </UFormGroup>
            <UFormGroup label="Genre" name="genere" v-bind="genreProps" class="space-y-4">
              <UInputMenu name="genre" v-model="genre" :options="genres" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50" />
            </UFormGroup>
            <UFormGroup :error="durationProps.error" class="relative space-y-4">
              <template #label>
                Duration: {{ durationString }}
                <div class="absolute top-0 right-0 flex gap-4">
                  <UButton
                    icon="i-heroicons-plus"
                    @click="updateDuration(5)"
                    :disabled="duration >= 14400"
                  />
                  <UButton
                    icon="i-heroicons-minus"
                    @click="updateDuration(-5)"
                    :disabled="duration <= 900"
                  />
                </div>
              </template>
              <template #default>
                <URange v-model="duration" color="blue" class="mt-6" :min="900" :max="14400" :step="5"/>
              </template>
            </UFormGroup>
            <div class="col-span-2 bg-gray-800 p-4 text-lg my-4 font-bold border-b border-b-gray-50">
              Media
            </div>
            <UFormGroup label="Thumbnails" :error="thumbnailsProps.error" help="The image must be less than 10mb and an image format">
              <UInput type="file" accept="image/*" name="thumbnail" id="thumbnail" @change="onFileChange" multiple/>
            </UFormGroup>
            <div class="col-span-2 bg-gray-800 p-4 text-lg my-4 font-bold border-b border-b-gray-50">
              Casts and Crews
            </div>
            <UFormGroup label="Director" :error="directorProps.error" class="space-y-4">
              <USelectMenu name="director" :options="directors" value-attribute="id" option-attribute="name" type="text" v-model="director" searchable/>
            </UFormGroup>
            
            <UFormGroup class="space-y-4" label="Actors" :error="actorsProps.error">
              <USelectMenu :options="directors" v-model="actors" value-attribute="id" option-attribute="name"  multiple searchable/>
            </UFormGroup>
            <div class="col-span-2 bg-gray-800 p-4 text-lg mt-4 mb-2 font-bold border-b border-b-gray-50">
              Description
            </div>
            <UFormGroup label="Discription" name="description" :error="descriptionProps.error" class="space-y-4 col-span-2">
              <UTextarea name="description" v-model="description" autoresize class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50" />
            </UFormGroup>
            <div class="col-span-2 bg-gray-800 p-4 text-lg mt-4 mb-2 font-bold border-b border-b-gray-50">
              Schedule
            </div>
            <TestScheduleForm name="schedules" v-model="schedules" @update:model-value="schedules = $event" />
          </div>
        </div>
        <button type="submit" @click="onSubmit" class="inline-flex justify-center py-2 px-4 border justify-self-end ml-auto border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
          Save Movie
        </button>
      </UForm>
    </div>
  </template>
  
<script setup lang="ts">
  import { useForm } from 'vee-validate'
  import * as yup from 'yup'
  import type { DurationReturn } from '~/utils/secondToString';
  import stos from '~/utils/secondToString';
  import VueDatePicker from '@vuepic/vue-datepicker';
  import '@vuepic/vue-datepicker/dist/main.css';
  
  const onFileChange = (files: FileList) => {
    if (files) {
      thumbnails.value = Array.from(files);
    }
  }

  const genres = [
    'Action', 'Comedy', 'Drama', 'Horror', 'Romance', 'Sci-Fi', 'Thriller'
  ]

  const directors = [
    {name: 'Abebe Chala', id: 1}, {name: 'Bekele Chala', id: 2}, {name: 'Debebe Chala', id: 3}
  ]

  const schema = yup.object().shape({
    title: yup.string().required('Title is required'),
    releaseDate: yup.date().required('Release date is required'),
    genre: yup.string().required('Genre is required').oneOf(genres),
    duration: yup.number().min(900, "Must be atleast 15 min").max(14400, "4 hours is enough").required("It is required to set a"),
    director: yup.number().required('You must choose at least one director'),
    actors: yup.array(yup.number()).min(1).required('You must choose at least one actors'),
    description: yup.string().required('Description is required'),
    thumbnails: yup.mixed().required('You need to select at least one thumbnail').test('fileSize', 'File size is too large', (value: any) => {
      if (!value) return true
      if (Array.isArray(value) && value.every((file) => file.size < (1 << 20) * 10)) return true
      return false
    })
  })
  
  const {defineField, handleSubmit, values } = useForm({
    validationSchema: schema,
  })

  const nuxtUiConfig = {
    props: (state: { errors: any[]; }) => ({
      error: state.errors[0]
    }),
  };

  const [title, titleProps] = defineField('title', nuxtUiConfig);
  const [releaseDate, releaseDateProps] = defineField('releaseDate', nuxtUiConfig);
  const [genre, genreProps] = defineField('genre', nuxtUiConfig);
  const [description, descriptionProps] = defineField('description', nuxtUiConfig);
  const [duration, durationProps] = defineField('duration', nuxtUiConfig);
  const [director, directorProps] = defineField('director', nuxtUiConfig);
  const [actors, actorsProps] = defineField('actors', nuxtUiConfig);
  const [thumbnails, thumbnailsProps] = defineField('thumbnails', nuxtUiConfig);
  const schedules = ref([])

  const updateDuration = (val: number) => {
    if (duration.value)
      duration.value += val
  }
  interface Duration {
    hours: number;
    minutes: number;
    seconds: number;
  }

  interface MovieFormValues {
    title: string;
    releaseDate: Date;
    genre: string;
    description: string;
    duration: Duration;
    director: string;
    thumbnails: File[];
  }

  const durationString: ComputedRef<string> = computed(()=>{
    const obj: DurationReturn = stos(duration.value)
    let s: string = ''
    if (obj.hours && obj.hours > 0)
      s += obj.hours.toString().padStart(2, '0') + ' Hours,   '
    if (obj.minutes && obj.minutes > 0)
      s += obj.minutes.toString().padStart(2, '0') + ' Min,   '
    if (obj.seconds && obj.seconds > 0)
      s += obj.seconds.toString().padStart(2, '0') + ' Seconds'
    return s
  });
  const onSubmit = handleSubmit((values) => {
    console.log("schedules: ", schedules)
    console.log("submit: ", values.thumbnails)
    // Here you would typically send the data to your backend
  })
</script>