<template>
  <div class="container mx-auto p-6 flex-1">
    <UButton icon="i-heroicons-backward" color="black" variant="outline" @click="useRouter().back()" class="mb-5" size="lg"></UButton>
    <h1 class="text-3xl font-bold mb-6">Edit general info of a movie</h1>
    <form @submit="onSubmit" method="post">
      <div class="p-6">
        <div class="col-span-2 bg-gray-800 p-4 mb-6 text-lg font-bold border-b border-b-gray-50">
          General info
        </div>
        <div class="grid grid-cols-2 gap-6">
          <UFormGroup label="Title" name="title" v-bind="titleProps" class="space-y-4">
            <UInput name="title" v-model="title" type="text" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50" />
          </UFormGroup>
          <UFormGroup label="Release Date" name="published_at" v-bind="releaseDateProps" class="space-y-4">
            <VueDatePicker name="published_at" v-model="releaseDate" :enable-time-picker="false" dark/>
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
                  :disabled="(duration || 0) >= 14400"
                />
                <UButton
                  icon="i-heroicons-minus"
                  @click="updateDuration(-5)"
                  :disabled="(duration || 0) <= 900"
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
            <UInput type="file" accept="image/*" name="thumbnail" id="thumbnail" @change="thumbnails=$event" multiple/>
          </UFormGroup> 
          <div class="flex justify-between gap-4 col-span-2">
            <div class="w-1/2">
              <p class="text-sm text-gray-500 mb-2">Preview of the selected image</p>
              <PreviewImage :files="thumbnails" />
            </div>
            <div class="w-1/2 border-l-2 border-primary-400 pl-3">
              <p class="text-sm text-gray-500 mb-2">Previous images</p>
              <ImageList :initial-images="images" :id="(id as string)" :featured="featured" @update:featured="onFeatured"/>
            </div>
          </div>
          <div class="col-span-2 bg-gray-800 p-4 text-lg mt-4 mb-2 font-bold border-b border-b-gray-50">
            Description
          </div>
          <UFormGroup label="Discription" name="description" :error="descriptionProps.error" class="space-y-4 col-span-2">
            <UTextarea name="description" v-model="description" autoresize class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50" />
          </UFormGroup>
        </div>
      </div>
      <button type="submit" class="inline-flex justify-center py-2 px-4 border justify-self-end ml-auto border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
        Save Movie
      </button>
    </form>
  </div>
</template>

<script setup lang="ts">
import { useForm } from 'vee-validate'
import * as yup from 'yup'
import type { DurationReturn } from '~/utils/secondToString';
import stos from '~/utils/secondToString';
import VueDatePicker from '@vuepic/vue-datepicker';
import '@vuepic/vue-datepicker/dist/main.css';
import { genres } from '~/constants';
import { MOVIE_BYID } from '~/graphql/queries/movies';
import type { Movie } from '~/types/movie';
import { MOVIE_THUMBNAIL_ADD } from '~/graphql/mutations/movie';

const id = useRoute().params.id
const { result } = useQuery<{movies_by_pk: Movie}>(MOVIE_BYID, { id });

const images = computed(()=> result.value?.movies_by_pk.movie_thumbnails?.map(({image_url}) => image_url) || [] )

const featured = ref(result.value?.movies_by_pk.featured_image || '')

const onFeatured = (image: string) => {
  console.log('image empty', image)
  console.log('prev', featured.value)
  console.log('result', result.value?.movies_by_pk.featured_image)
  if (image === '')
    featured.value = result.value?.movies_by_pk.featured_image || ''
  else
    featured.value = image
  console.log('result after', featured.value)
}
const schema = yup.object().shape({
  title: yup.string().required('Title is required'),
  published_at: yup.date().required('Release date is required'),
  genre: yup.string().required('Genre is required').oneOf(genres),
  duration: yup.number().min(900, "Must be atleast 15 min").max(14400, "4 hours is enough").required("It is required to set a"),
  description: yup.string().required('Description is required'),
  thumbnails: yup.mixed().nullable().test('fileSize', 'File size is too large', (value: any) => {
    if (!value) return true
    value = Array.from(value)
    if (Array.isArray(value) && value.every((file) => file.size < (1 << 20) * 10)) return true
    return false
  })
})

const toast = useToast()
const initial = ref({
  title: result.value?.movies_by_pk.title,
  published_at: result.value?.movies_by_pk.published_at,
  genre: result.value?.movies_by_pk.genre?.toLocaleUpperCase(),
  duration: (result.value?.movies_by_pk.duration),
  description: (result.value?.movies_by_pk.description),
  thumbnails: null
})

const {defineField, handleSubmit, values } = useForm({
  validationSchema: schema,
  initialValues: initial.value
})

const nuxtUiConfig = {
  props: (state: { errors: any[]; }) => ({
    error: state.errors[0]
  }),
};

const [title, titleProps] = defineField('title', nuxtUiConfig);
const [releaseDate, releaseDateProps] = defineField('published_at', nuxtUiConfig);
const [genre, genreProps] = defineField('genre', nuxtUiConfig);
const [description, descriptionProps] = defineField('description', nuxtUiConfig);
const [duration, durationProps] = defineField('duration', nuxtUiConfig);
const [thumbnails, thumbnailsProps] = defineField('thumbnails', nuxtUiConfig);

const updateDuration = (val: number) => {
  if (duration.value)
    duration.value += val
}

const durationString: ComputedRef<string> = computed(()=>{
  const obj: DurationReturn = stos(duration.value || 0)
  let s: string = ''
  if (obj.hours && obj.hours > 0)
    s += obj.hours.toString().padStart(2, '0') + ' Hours,   '
  if (obj.minutes && obj.minutes > 0)
    s += obj.minutes.toString().padStart(2, '0') + ' Min,   '
  if (obj.seconds && obj.seconds > 0)
    s += obj.seconds.toString().padStart(2, '0') + ' Seconds'
  return s
});

console.log("Initial values:", initial.value)
console.log("Initial values:", values)
const {executeUpdate, loading, onDone} = useUpdateMovie()
onDone((data) => {
  toast.add({
    color: 'green',
    title: 'Movie Updated',
    description: 'The movie has been updated successfully',
  })
  useRouter().replace(`/movies/${id}`)
})

const {mutate: inserThumbnails, onDone: imagesRegistered} = useMutation(MOVIE_THUMBNAIL_ADD)

imagesRegistered((res) => {
  console.log("Images registered", res)
})

const onSubmit = handleSubmit(async (values) => {
  console.log("Values:", values)
  try {
    console.log("SUBBMITINEG")
    if (values.thumbnails && (values.thumbnails as FileList).length > 0) {
      const { data, status } = await useUploadImages(values.thumbnails);
      if (status.value !== 'success' || !data.value) return
      if (!data.value) return
      const castedData = Array.isArray(data.value) ? data.value.map(item => ({ movie_id: id as string, image_url: item.image_url })) : [];
      console.log("Casted data:", castedData)
      await inserThumbnails({objects: castedData})
    }

    const object = {
      title: values.title,
      published_at: values.published_at,
      genre: (values.genre as string).toLowerCase().trim(),
      duration: values.duration,
      description: values.description,
      featured_image: featured.value
    }
    await executeUpdate(id as string, object)
  } catch (error) {
    console.log("Error casting data", error)
  }
})

</script>