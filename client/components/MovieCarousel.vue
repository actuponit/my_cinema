<template>
  <div class=" text-white">
    <div class="container mx-auto px-4 py-10">
      <h2 class="text-4xl font-bold mb-8 text-center">This Week's Top Cinema Schedules</h2>
      <div class="relative">
        <div class="overflow-hidden">
          <div v-if="movies && movies.length > 0" class="flex transition-transform duration-500 ease-in-out"
            :style="{ transform: `translateX(-${currentSlide * 100}%)` }">
            <div v-for="(movie, index) in movies" :key="index" class="w-full flex-shrink-0 px-4">
              <div class="bg-gray-800 rounded-lg overflow-hidden shadow-xl">
                <div class="lg:flex">
                  <div class="lg:flex-shrink-0">
                    <img :src="movie.image" :alt="movie.title" class="h-48 w-full object-cover lg:h-full lg:w-80">
                  </div>
                  <div class="lg:p-8 p-2 flex-1">
                    <div class="uppercase tracking-wide text-sm text-primary font-semibold">{{ movie.genre }}</div>
                    <h3 class="mt-1 text-2xl font-bold leading-tight">{{ movie.title }}</h3>
                    <p class="mt-2 text-gray-400">{{ movie.description }}</p>
                    <span class="flex gap-5">
                      <p class="mt-2 text-gray-400">Ratings: {{ movie.rating }}/5</p>
                      <p class="mt-2 text-gray-400">Total Reviews: {{ movie.reviews }}</p>
                    </span>
                    <div class="mt-4">
                      <span class="text-gray-100 font-bold">Showtimes:</span>
                      <ul class="mt-2 space-y-1">
                        <li v-for="(time, timeIndex) in movie.showtimes" :key="timeIndex"
                          class="text-gray-400 flex justify-between items-center border-b-gray-700 border-b-2 py-1">
                          {{ formatDateFull(time.start_time) }}
                          <UButton size="sm" @click="bookTicket(time)">Book Now</UButton>
                        </li>
                      </ul>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
        <button @click="prevSlide"
          class="absolute left-0 top-1/2 transform -translate-y-1/2 bg-gray-700 bg-opacity-50 hover:bg-opacity-75 rounded-r-lg px-2 py-4 focus:outline-none"
          aria-label="Previous slide">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
          </svg>
        </button>
        <button @click="nextSlide"
          class="absolute right-0 top-1/2 transform -translate-y-1/2 bg-gray-700 bg-opacity-50 hover:bg-opacity-75 rounded-l-lg px-2 py-4 focus:outline-none"
          aria-label="Next slide">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
        </button>
      </div>
      <div class="flex justify-center mt-4 space-x-2">
        <button v-for="(_, index) in movies" :key="index" @click="goToSlide(index)" :class="[
          'w-3 h-3 rounded-full focus:outline-none',
          currentSlide === index ? 'bg-primary' : 'bg-gray-600 hover:bg-gray-500'
        ]" :aria-label="`Go to slide ${index + 1}`"></button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { HOME_QUERY } from '~/graphql/queries/home';
import BookingModal from './BookingModal.vue';

const today = new Date();
const nextweek = new Date(today);
nextweek.setDate(today.getDate() + 7);
const { result, loading, refetch } = useQuery(HOME_QUERY, { gte: today.toISOString(), lte: nextweek.toISOString() });
await refetch();
const currentSlide = ref(0);

const modal = useModal();
function openModal (schedule: any) {
	modal.open(BookingModal, {
    movieDetails: schedule,
    onClose () {
      modal.close()
    }
  })
}

const bookTicket = (schedule: any) => {
  console.log("schedule", schedule)
	openModal({start: formatDateShort(schedule.start_time), id: schedule.id, time: formatTime(schedule.start_time), hall: schedule.hall, format: cinemaFormatReverse(schedule.format), price: schedule.price})
}
const movies = computed(() => {
  return result.value?.movies.map((s: any) => ({
    title: s.title,
    genre: s.genre,
    rating: s.average_rating,
    reviews: s.total_rating,
    description: s.description,
    image: displayImage(s.featured_image),
    showtimes: s.schedules,
  })) || [];
});
console.log("movies", movies.value)

const nextSlide = () => {
  currentSlide.value = (currentSlide.value + 1) % movies.value.length;
};

const prevSlide = () => {
  currentSlide.value = (currentSlide.value - 1 + movies.value.length) % movies.value.length;
};

const goToSlide = (index:number) => {
  currentSlide.value = index;
};
</script>
