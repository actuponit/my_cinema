<template>
  <div class="min-h-screen p-8 flex-1 max-w-full">
      <div class="rounded-lg overflow-hidde shadow-inner shadow-blue-300">
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
                <img :src="item" :alt="movie.title" class="w-full h-96 object-fill object-center rounded-s-lg" />
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

							<UButton v-if="user?.role !== 'cinema'" size="lg" variant="ghost" :icon="movie.is_bookmarked?'i-heroicons-bookmark-solid':'i-heroicons-bookmark'" :label="movie.is_bookmarked?'Bookmarked':'Bookmark'" @click="onBookMark"/>
              <div v-if="user?.role === 'cinema'">
                <UButton size="lg" variant="ghost" icon="i-heroicons-pencil" label="Edit" @click="onEdit"/>
                <UButton size="lg" color="red" variant="ghost" icon="i-heroicons-trash" label="Delete" @click="onDeleteMovie"/>
              </div>
						</div>
            <div class="flex items-center mb-4">
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
              <p><span class="font-semibold">Release Date:</span> {{ formatDateShort(movie.releaseDate) }}</p>
            </div>
          </div>
        </div>
        
        <!-- Cast and Crew -->
        <div class="p-6 border-t border-gray-700 w-[calc(100vw-20rem)] overflow-x-scroll">
          <div class="flex w-full items-center">
            <h2 class="text-2xl flex-1 font-bold text-white mb-4">Cast and Crew</h2>
            <UButton v-if="user?.role === 'cinema'" size="sm" variant="link" color="primary" class="flex-shrink-0" @click="openChangeDirectorModal">Change director</UButton>
          </div>
          <div class="flex gap-8 flex-nowrap">
            <div v-for="person in movie.castAndCrew" :key="person.id" class="text-center min-w-fit relative">
              <UAvatar :src="person.image" size="xl" class="mb-2" />
              <p class="font-semibold text-white mx-2">{{ person.name }}</p>
              <p class="text-sm text-gray-400">{{ person.role }}</p>
              <UButton size="xs" icon="i-heroicons-trash" v-if="person.role !== 'Director' && user?.role === 'cinema'" color="red" variant="ghost" class="absolute top-0 right-0" @click="openRemoveConfirmModal(person)" />
            </div>
            <div v-if="user?.role === 'cinema'" class="text-center min-w-fit flex flex-col justify-center items-center cursor-pointer" @click="openAddActorModal">
              <UAvatar icon="i-heroicons-plus-circle" size="xl" class="mb-2"/>
              <UButton variant="link">Add actor</UButton>
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
                <StarRating :initial-rating="review.rating" :read-only="true"/>
              </div>
              <p class="text-gray-300">{{ review.comment }}</p>
            </div>
          </div>
          <div v-else>
            <ErrorComponent title="No reviews yet" message="No reveiws for this movie yet" icon="i-heroicons-archive-box-x-mark"/>
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
          <div class="flex items-start">
            <h2 class="text-2xl flex-1 font-bold text-white mb-4">Upcoming Schedules</h2>
            <UButton v-if="user?.role === 'cinema'" variant="link" @click="openAddScheduleModal">Add schedule</UButton>
          </div>
          <UTable :columns="scheduleColumns" :rows="movie.schedules" >
            <template #actions-data="{ row }">
              <UButton v-if="user?.role === 'cinema'" size="sm" color="red" @click="openDeleteScheduleModal(row)">Delete schedule</UButton>
              <UButton v-else size="sm" @click="bookTicket(row)">Book Now</UButton>
            </template>
          </UTable>
        </div>
      </div>
  </div>
</template>

<script setup lang="ts">
import { AddScheduleModal, BookingModal } from '#components';
import { offset } from '@popperjs/core';
import AddActorModal from '~/components/AddActorModal.vue';
import ChangeDirectorModal from '~/components/ChangeDirectorModal.vue';
import ConfirmModal from '~/components/ConfirmModal.vue';
import { DELTE_MOVIE } from '~/graphql/mutations/movie';
import { DELETE_SCHEDULE } from '~/graphql/mutations/schedule';
import { MOVIE_BYID, MOVIES_QUERY } from '~/graphql/queries/movies';
import type { Movie, Schedule } from '~/types/movie';

const id = useRoute().params.id
const modal = useModal()
const toast = useToast();
const { user } = useUser();

// Mock data for the movie
const variables = computed(() => (user.value && user.value.role === 'cinema')?({id}):({id, user_id: user.value?.id}))
const { result, loading } = useQuery<{movies_by_pk: Movie}>(MOVIE_BYID, variables);
console.log(result.value)
const onEdit = () => {
  useRouter().push('/admin/movies/create/'+(id as string))
}

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
},)

const reviews = computed(() => {
  const r = result.value?.movies_by_pk.ratings?.map(({feedback, userByUser, rating}) => ({
    id: userByUser?.id,
    userName: userByUser?.first_name + ' ' + userByUser?.last_name,
    rating,
    comment: feedback,
  })) || [];
  return showAll.value ? r : r.slice(0, 2)
})

const schedules = computed(() => {
  return result.value?.movies_by_pk.schedules?.map(({id, start_time, hall, format, price}) => ({
    id,
    start: formatDateShort(start_time || ''),
    time: formatTime(start_time || ''),
    hall: hall,
    format: cinemaFormatReverse(format || ''),
    price: price,
  })) || []
})

const images = computed(() => {
  return result.value?.movies_by_pk.movie_thumbnails?.map(({image_url}) => displayImage(image_url)) || []
})


// const review = ref('')
const movie = computed(() => ({
  id: id,
  title: result.value?.movies_by_pk.title,
  images: images.value,
  rating: result.value?.movies_by_pk.average_rating,
  description: result.value?.movies_by_pk.description,
  genre: result.value?.movies_by_pk.genre,
  duration: secondToString(result.value?.movies_by_pk.duration || 0),
  releaseDate: formatDateShort(result.value?.movies_by_pk.published_at || ''),
  castAndCrew: castAndCrew.value,
  reviews: reviews.value,
  schedules: schedules.value,
  is_bookmarked: result.value?.movies_by_pk.is_bookmarked || false
}))

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

// Function to submit a review
const { executeInsert, onDone } = useInsertReview(id as string)

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
  if (!(user.value && user.value?.id)) {
    toast.add({
      title: "Not logged in",
      description: "You need to login to be able to give reviews",
      color: "red"
    })
    return
  }
  const values = {
    rating: rating.value,
    feedback: comment.value,
    movie: id,
    user: user.value?.id
  }
  await executeInsert(values)
}
const showAll = ref(false)

const openAddActorModal = () => {
  modal.open(AddActorModal, {
    id: id as string,
    onClose: () => {
      modal.close()
    }
  })
}

const openChangeDirectorModal = () => {
  modal.open(ChangeDirectorModal, {
    id: id as string,
    onClose: () => {
      modal.close()
    }
  })
}

const {executeDelete, onDone: deletedActor} = useCastDeleteMovie()
deletedActor((res) => {
  toast.add({
    title: "Successful Operation",
    description: "Successfully removed actor from crew!",
    color: 'green'
  })
  console.log("Deleted actor", res)
  modal.close()
})

function openRemoveConfirmModal (person: any) {
  modal.open(ConfirmModal, {
    title: 'Remove Actor',
    message: `Are you sure you want to remove ${person.name} from the crew?`,
    action: 'Remove',
    async onAction() {
      console.log('Removing crew member:', person.id)
      await executeDelete(person.id, id as string)
    },
    onClose () {
      modal.close()
    }
  })
}

function openAddScheduleModal () {
  console.log("here")
  modal.open(AddScheduleModal, {
    id: id as string,
    onClose() {
      modal.close()
    }
  })
}

function openModal (schedule: Schedule) {
	modal.open(BookingModal, {
    movieDetails: schedule,
    onClose () {
      modal.close()
    }
  })
}

const {mutate, onDone: onDeletedSchedule} = useMutation(DELETE_SCHEDULE, {
  refetchQueries: [{query: MOVIE_BYID, variables: {id}}]
});
onDeletedSchedule(() => {
  toast.add({
    title: "Successful Operation",
    description: "Successfully deleted schedule!",
    color: 'green'
  })
  console.log("Deleted schedule")
  modal.close()
})

function openDeleteScheduleModal (sch: any) {
  modal.open(ConfirmModal, {
    title: 'Remove Schedule',
    message: `Are you sure you want to remove schedule at ${sch.start} ${sch.time} at venue ${sch.hall}?`,
    action: 'Remove',
    async onAction() {
      await mutate({id: sch.id})
    },
    onClose () {
      modal.close()
    }
  })
}
// Function to book a ticket
const bookTicket = (schedule: any) => {
	openModal(schedule)
}

const {executeInsert:bookmark, onDone:bookmarked, loading:bookmarking} = useBookMark();

bookmarked(() => {
  toast.add({
    title: "Succesfull operation",
    description: "Successfully added movie to your bookmarks",
    color: "green"
  });
})
const onBookMark = async () => {
  if (!user.value) {
    toast.add({
      title: "Can't bookmark",
      description: "You need to login to bookmark movies",
      color: "red"
    });
    return
  }
  await bookmark({
    movie_id: id,
    user: user.value.id
  })
}

const {mutate: deleteMovie, onDone: onDeletedMovie} = useMutation(DELTE_MOVIE);

onDeletedMovie(() => {
  toast.add({
    title: "Successful Operation",
    description: "Successfully deleted movie!",
    color: 'green'
  })
  modal.close()
  useRouter().replace('/movies')
})
const onDeleteMovie = () => {
  console.log("Deleting movie", id)
  modal.open(ConfirmModal, {
    title: 'Delete Movie',
    message: `Are you sure you want to delete this movie?`,
    action: 'Delete',
    async onAction() {
      await deleteMovie({id}, {
  refetchQueries: [{query: MOVIES_QUERY, variables: { where: {}, offset: 0 }}]
})
    },
    onClose () {
      modal.close()
    }
  })
}
</script>
