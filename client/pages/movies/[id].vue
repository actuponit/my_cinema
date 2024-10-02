<template>
  <div class="min-h-screen p-8 flex-1 max-w-full">
      <div class="rounded-lg shadow-xl overflow-hidden shadow-black-50/20">
        <div class="md:flex">
          <!-- Image Carousel -->
          <div class="md:w-1/2">
            <UCarousel 
                :items="movie.images" 
                :ui="{ 
                    item: 'basis-full',
                    indicators: {
                        wrapper: 'absolute bottom-0 bg-black-950 bg-opacity-70 p-2',
                    }
                }" 
                :prev-button="{
                    color: 'gray',
                    icon: 'i-heroicons-arrow-left-20-solid',
                }"
                :next-button="{
                    color: 'gray',
                    icon: 'i-heroicons-arrow-right-20-solid',
                }"
                arrows
                indicators
            >
              <template v-slot:default="{ item }">
                <img :src="item" :alt="movie.title" class="w-full h-96 object-fill object-center" />
              </template>
            
              <template v-slot:indicator="{ onClick, page, active }">
                <UButton :label="String(page)" :variant="active ? 'solid' : 'outline'" size="2xs" class="rounded-full min-w-6 justify-center" @click="onClick(page)" />
              </template>
            </UCarousel>
          </div>
          
          <!-- Movie Details -->
          <div class="md:w-1/2 p-6">
						<div class="flex justify-between item-center">
							<h1 class="text-3xl font-bold text-white mb-2">{{ movie.title }}</h1>
							<UButton size="lg" variant="ghost" icon="i-heroicons-bookmark" label="Bookmark"/>
						</div>
            <div class="flex items-center mb-4">
              <URating v-model="movie.rating" :length="5" />
              <span class="ml-2 text-sm text-gray-400">{{ movie.rating?.toFixed(1) }} / 5</span>
            </div>
            <p class="text-gray-300 mb-4">{{ movie.description }}</p>
            <div class="space-y-2">
              <p><span class="font-semibold">Genre:</span> {{ movie.genre }}</p>
              <p><span class="font-semibold">Duration:</span> 
                {{ movie.duration.hours > 0 ? movie.duration.hours + ' Hours, ':'' }} 
                {{ movie.duration.hours > 0 ? movie.duration.minutes + ' Mins, ':'' }} 
                {{ movie.duration.hours > 0 ? movie.duration.seconds + ' Seconds':'' }} 
              </p>
              <p><span class="font-semibold">Release Date:</span> {{ formatDate(movie.releaseDate) }}</p>
            </div>
          </div>
        </div>
        
        <!-- Cast and Crew -->
        <div class="p-6 border-t border-gray-700 w-[calc(100vw-20rem)] overflow-x-scroll">
          <h2 class="text-2xl font-bold text-white mb-4">Cast and Crew</h2>
          <div class="flex gap-8 flex-nowrap">
            <div v-for="person in movie.castAndCrew" :key="person.id" class="text-center min-w-fit">
              <UAvatar :src="person.image" size="xl" class="mb-2" />
              <p class="font-semibold text-white mx-2">{{ person.name }}</p>
              <p class="text-sm text-gray-400">{{ person.role }}</p>
            </div>
          </div>
        </div>
        
        <!-- Reviews -->
        <div class="p-6 border-t border-gray-700">
          <div class="flex justify-between">
              <h2 class="text-2xl font-bold text-white mb-4">Reviews</h2>
              <p class="text-lg text-primary cursor-pointer underline underline-offset-2" @click="showAll = !showAll" >Show all reviews</p>
          </div>
          <div v-if="reviews.length > 0" class="space-y-4 mb-6">
            <div v-for="review in reviews" :key="review.id" class="bg-gray-700 p-4 rounded-lg">
              <div class="flex items-center mb-2 w-full justify-between">
                <span class="font-semibold text-white">{{ review.userName }}</span>
                <StarRating :initial-rating="3.75" :read-only="true"/>
              </div>
              <p class="text-gray-300">{{ review.comment }}</p>
            </div>
          </div>
          <div v-else>
            <ErrorComponent title="No reviews yet" icon="i-heroicons-archive-box-x-mark"/>
          </div>
          
          <!-- Add Review Form -->
          <UCard>
            <template #header>
              <h3 class="text-lg font-semibold">Add Your Review</h3>
            </template>
            <form @submit="submitReview">
              <UFormGroup label="Rating" name="rating">
                <StarRating v-model:rating="rating" />
              </UFormGroup>
              <UFormGroup label="Comment" name="comment" class="mt-4">
                <UTextarea v-model="comment" />
              </UFormGroup>
              <UButton type="submit" color="primary" class="mt-4">Submit Review</UButton>
            </form>
          </UCard>
        </div>
        
        <!-- Upcoming Schedules -->
        <div class="p-6 border-t border-gray-700">
          <h2 class="text-2xl font-bold text-white mb-4">Upcoming Schedules</h2>
          <UTable :columns="scheduleColumns" :rows="movie.schedules">
            <template #actions-data="{ row }">
              <UButton size="sm" @click="bookTicket(row)">Book Now</UButton>
            </template>
          </UTable>
        </div>
      </div>
  </div>
</template>

<script setup lang="ts">
import { BookingModal } from '#components';
import { MOVIE_BYID } from '~/graphql/queries/movies';
import type { ReviewForm } from '~/types';
import type { Movie, Schedule } from '~/types/movie';

const id = useRoute().params.id
// Mock data for the movie
console.log("id", id)
const { result } = useQuery<{movies_by_pk: Movie}>(MOVIE_BYID, { id });
console.log("result", result.value)
const castAndCrew = computed(() => {
  const actors = result.value?.movies_by_pk.crews?.map(({cast}) => ({
    id: cast.id,
    name: cast.first_name + ' ' + cast.last_name,
    role: 'Actor',
    image: displayImage(cast.photo_url)
  }))
  const director = {
    id: result.value?.movies_by_pk.my_director?.id,
    name: result.value?.movies_by_pk.my_director?.first_name + ' ' + result.value?.movies_by_pk.my_director?.last_name,
    role: 'Director',
    image: displayImage(result.value?.movies_by_pk.my_director?.photo_url)
  }
  return [director, ...(actors || [])]
})
// const review = ref('')
const movie = reactive({
  id: 1,
  title: result.value?.movies_by_pk.title,
  images: result.value?.movies_by_pk.movie_thumbnails?.map(({image_url}) => displayImage(image_url)),
  rating: result.value?.movies_by_pk.average_rating,
  description: result.value?.movies_by_pk.description,
  genre: result.value?.movies_by_pk.genre,
  duration: secondToString(result.value?.movies_by_pk.duration || 0),
  releaseDate: formatDateShort(result.value?.movies_by_pk.published_at || ''),
  castAndCrew: castAndCrew.value,
  reviews: result.value?.movies_by_pk.ratings?.map(({feedback, userByUser, rating}) => ({
    id: userByUser?.id,
    userName: userByUser?.first_name + ' ' + userByUser?.last_name,
    rating,
    comment: feedback,
  })) || [],
  schedules: result.value?.movies_by_pk.schedules?.map(({id, start_time, hall, format, price}) => ({
    id,
    start: formatDateShort(start_time || ''),
    time: `${formatTime(start_time || '')}`,
    hall: hall,
    format: cinemaFormatReverse(format || ''),
    price: price,
  })) || [],
})

// Form state for adding a review
const rating = ref(0)
const comment = ref('')

// Columns for the schedule table
const scheduleColumns = [
  { key: 'start', label: 'Date' },
  { key: 'time', label: 'Time' },
  { key: 'hall', label: 'Cinema Hall' },
  { key: 'format', label: 'Format' },
  { key: 'price', label: 'Price' },
  { key: 'actions', label: 'Book' },
]

// Function to format date
const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' })
}

// Function to submit a review
const { executeInsert, onDone } = useInsertReview(id as string)
const toast = useToast();
const router = useRouter();
onDone(() => {
  toast.add({
    title: "Successful Operation",
    description: "Successfully submited reveiw!",
    color: 'green'
  })
})
const submitReview = async (event: Event) => {
  // console.log('Submitting review:', event.data)
  event.preventDefault();
  const {user} = useUser();
  if (!(user.value && user.value.id)) {
    toast.add({
      title: "Not logged in",
      description: "You need to login to be able to give reviews",
      color: "red"
    })
    return
  }
  console.log("ID: ", user.value.id)
  const values = {
    rating: rating.value,
    feedback: comment.value,
    movie: id,
    user: user.value.id
  }
  await executeInsert(values)
}
const showAll = ref(false)
const reviews = computed(() => {
  return showAll.value ? movie.reviews : movie.reviews.slice(0, 2)
})

const modal = useModal()

function openModal (schedule: Schedule) {
	modal.open(BookingModal, {
    movieDetails: schedule,
    onClose () {
      modal.close()
    }
  })
}
// Function to book a ticket
const bookTicket = (schedule: Schedule) => {
	openModal(schedule)
  console.log('Booking ticket for schedule:', schedule)
}
</script>
