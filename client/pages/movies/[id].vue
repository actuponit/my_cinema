<template>
  <div class="min-h-screen flex-1 bg-black-950">
    <UContainer class="py-8">
      <div class="bg-gray-800 rounded-lg shadow-xl overflow-hidden">
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
              <span class="ml-2 text-sm text-gray-400">{{ movie.rating.toFixed(1) }} / 5</span>
            </div>
            <p class="text-gray-300 mb-4">{{ movie.description }}</p>
            <div class="space-y-2">
              <p><span class="font-semibold">Genre:</span> {{ movie.genre.join(', ') }}</p>
              <p><span class="font-semibold">Duration:</span> {{ movie.duration }}</p>
              <p><span class="font-semibold">Release Date:</span> {{ formatDate(movie.releaseDate) }}</p>
            </div>
          </div>
        </div>
        
        <!-- Cast and Crew -->
        <div class="p-6 border-t border-gray-700 overflow-x-scroll scrollbar-thin">
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
          <div class="space-y-4 mb-6">
            <div v-for="review in reviews" :key="review.id" class="bg-gray-700 p-4 rounded-lg">
              <div class="flex items-center mb-2 w-full justify-between">
                <span class="font-semibold text-white">{{ review.userName }}</span>
                <StarRating :initial-rating="3.75" :read-only="true"/>
              </div>
              <p class="text-gray-300">{{ review.comment }}</p>
            </div>
          </div>
          
          <!-- Add Review Form -->
          <UCard>
            <template #header>
              <h3 class="text-lg font-semibold">Add Your Review</h3>
            </template>
            <UForm :state="reviewForm" @submit="submitReview">
              <UFormGroup label="Rating" name="rating">
                <StarRating v-model:rating="rating" />
              </UFormGroup>
              <UFormGroup label="Comment" name="comment">
                <UTextarea v-model="reviewForm.comment" />
              </UFormGroup>
              <UButton type="submit" color="primary" class="mt-4">Submit Review</UButton>
            </UForm>
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
    </UContainer>
  </div>
</template>

<script setup lang="ts">
import { BookingModal } from '#components';
import type { ReviewForm, Schedule,  } from '~/types';

// import {BookingModal} from '#components';

// Mock data for the movie
const movie = ref({
  id: 1,
  title: 'Inception',
  images: [
    '/placeholder.webp',
    '/placeholder.webp',
    '/placeholder.webp',
  ],
  rating: 4.5,
  description: 'A thief who steals corporate secrets through the use of dream-sharing technology is given the inverse task of planting an idea into the mind of a C.E.O.',
  genre: ['Action', 'Adventure', 'Sci-Fi'],
  duration: '2h 28min',
  releaseDate: '2010-07-16',
  castAndCrew: [
    { id: 1, name: 'Christopher Nolan', role: 'Director', image: '/images/nolan.jpg' },
    { id: 2, name: 'Leonardo DiCaprio', role: 'Cobb', image: '/images/dicaprio.jpg' },
    { id: 3, name: 'Joseph Gordon-Levitt', role: 'Arthur', image: '/images/gordon-levitt.jpg' },
    { id: 4, name: 'Ellen Page', role: 'Ariadne', image: '/blank-profile-picture.webp' },
    { id: 4, name: 'Ellen Page', role: 'Ariadne', image: '/blank-profile-picture.webp' },
    { id: 4, name: 'Ellen Page', role: 'Ariadne', image: '/blank-profile-picture.webp' },
    { id: 4, name: 'Ellen Page', role: 'Ariadne', image: '/blank-profile-picture.webp' },
    { id: 4, name: 'Ellen Page', role: 'Ariadne', image: '/blank-profile-picture.webp' },
    { id: 4, name: 'Ellen Page', role: 'Ariadne', image: '/blank-profile-picture.webp' },
    { id: 4, name: 'Ellen Page', role: 'Ariadne', image: '/blank-profile-picture.webp' },
    { id: 4, name: 'Ellen Page', role: 'Ariadne', image: '/blank-profile-picture.webp' },
    { id: 4, name: 'Ellen Page', role: 'Ariadne', image: '/blank-profile-picture.webp' },
    { id: 4, name: 'Ellen Page', role: 'Ariadne', image: '/blank-profile-picture.webp' },
    { id: 4, name: 'Ellen Page', role: 'Ariadne', image: '/blank-profile-picture.webp' },
    { id: 4, name: 'Ellen Page', role: 'Ariadne', image: '/blank-profile-picture.webp' },
    { id: 4, name: 'Ellen Page', role: 'Ariadne', image: '/blank-profile-picture.webp' },
  ],
  reviews: [
    { id: 1, userName: 'John Doe', userImage: '/images/user1.jpg', rating: 5, comment: 'Mind-bending and visually stunning!' },
    { id: 2, userName: 'Jane Smith', userImage: '/images/user2.jpg', rating: 4, comment: 'Great concept, but a bit confusing at times.' },
    { id: 3, userName: 'Jane Smith', userImage: '/images/user2.jpg', rating: 4, comment: 'Great concept, but a bit confusing at times.' },
    { id: 4, userName: 'Jane Smith', userImage: '/images/user2.jpg', rating: 4, comment: 'Great concept, but a bit confusing at times.' },
  ],
  schedules: [
    { id: 1, date: '2023-06-15', time: '18:30', hall: 'Hall A', format: '2D', price: 12.50 },
    { id: 2, date: '2023-06-15', time: '21:00', hall: 'Hall B', format: '3D', price: 12.50 },
    { id: 3, date: '2023-06-16', time: '19:00', hall: 'Hall A', format: '2D', price: 12.50 },
    { id: 4, date: '2023-06-16', time: '22:00', hall: 'Hall C', format: 'IMAX', price: 12.50 },
  ],
})

// Form state for adding a review
const reviewForm = ref({
  rating: 0,
  comment: '',
})

// Define the rating property
const rating = ref(0)

// Columns for the schedule table
const scheduleColumns = [
  { key: 'date', label: 'Date' },
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

const submitReview = (event: { data: ReviewForm }) => {
  // Here you would typically send the review to your backend
  console.log('Submitting review:', event.data)
  // For this example, we'll just add it to the reviews array
  movie.value.reviews.push({
    id: movie.value.reviews.length + 1,
    userName: 'Current User',
    userImage: '/images/current-user.jpg',
    rating: event.data.rating,
    comment: event.data.comment,
  })
  // Reset the form
  reviewForm.value = { rating: 0, comment: '' }
}
const showAll = ref(false)
const reviews = computed(() => {
  return showAll.value ? movie.value.reviews : movie.value.reviews.slice(0, 2)
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
