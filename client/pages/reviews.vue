<template>
  <div class="min-h-screen text-gray-100 flex-1">
    <div class="container mx-auto px-4 py-8">
      <h1 class="text-4xl font-bold mb-8 text-center">Your Reviews</h1>

      <div v-if="reviews.length > 0">
        <ul class="grid gap-6 lg:grid-cols-3">
          <li v-for="review in reviews" :key="review.id"
            class="bg-gray-800 rounded-lg p-6 shadow-lg transition-all duration-300 hover:shadow-xl hover:scale-105">
            <div class="flex justify-between items-start mb-4">
              <h3 class="text-xl font-semibold">{{ review.title }}</h3>
              <UBadge color="yellow" size="lg">
                {{ review.rating }}/5
              </UBadge>
            </div>
            <p class="text-gray-300 mb-4">{{ review.feedback }}</p>
            <div class="flex justify-between items-center text-sm text-gray-400">
              <span>{{ formatDateFull(review.date) }}</span>
              <div>
                <UButton color="red" variant="ghost" icon="i-heroicons-trash" @click="onDelete(review.id)">
                  Delete
                </UButton>
                <UButton color="primary" variant="ghost" icon="i-heroicons-pencil-square" @click="onEdit(review)">
                  Edit
                </UButton>
              </div>
            </div>
          </li>
        </ul>
      </div>
      <ErrorComponent v-else icon="i-heroicons-chat-bubble-bottom-center-text" title="No reviews yet" />
    </div>
  </div>
</template>

<script setup lang="ts">
import EditReviewModal from '~/components/EditReviewModal.vue';
import { DELETE_REVIEW } from '~/graphql/mutations/movie';
import { MOVIE_BYID, MYREVIEWS_QUERY } from '~/graphql/queries/movies';
import { formatDateFull } from '~/utils/dateformat'

const { user } = useUser()
const { result } = useQuery(MYREVIEWS_QUERY, { _eq: user.value?.id }, { enabled: !!user.value })
const reviews = computed(() => {
  return result.value?.ratings.map((review: any) => {
    return {
      id: review.movieByMovie?.id,
      title: review.movieByMovie?.title,
      feedback: review.feedback,
      rating: review.rating,
      date: review.updated_at,
      editable: false
    }
  }) || []
})
const modal = useModal();
const onEdit = (review: any) => {
  modal.open(EditReviewModal,
    { review, user_id: user.value?.id, onClose() { modal.close() } }
  )
}

const {mutate, onDone} = useMutation(DELETE_REVIEW)
onDone(() => {
  useToast().add({
    title: 'Review deleted',
    description: 'Your review has been deleted successfully',
    color: 'green'
  })
})
const  onDelete = (id: string) => {
  console.log("id--d: : :", id)
  mutate({_eq: id}, {refetchQueries: [{query: MYREVIEWS_QUERY, variables: {_eq: user.value?.id}}, {query: MOVIE_BYID, variables: {id, user_id: user.value?.id}}]});
}

definePageMeta({
  title: 'My Reviews',
  middleware: ['must-login']
})
</script>