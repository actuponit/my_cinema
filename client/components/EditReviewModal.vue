<template>
  <UModal :ui="{ divide: 'divide-y divide-gray-100 dark:divide-gray-800' }">
    <UCard>
      <template #header>
        <div class="flex justify-between items-center">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
            Edit Review for {{ review.title }}
          </h3>
          <UButton color="gray" variant="ghost" icon="i-heroicons-x-mark-20-solid" @click="closeModal" />
        </div>
      </template>
        <UFormGroup label="Rating">
          <StarRating v-model:rating="rating" :initial-rating="rating"/>
        </UFormGroup>
        <UFormGroup label="Feedback" class="mt-8">
          <UTextarea v-model="feedback"/>
        </UFormGroup>
      <template #footer>
        <div class="flex justify-end space-x-3">
          <UButton color="gray" variant="outline" @click="closeModal">
            Cancel
          </UButton>
          <UButton color="primary" @click="onEdit">
            Edit Review
          </UButton>
        </div>
      </template>
    </UCard>
  </UModal>
</template>
<script setup lang="ts">
// import { EDIT_REVIEW } from '~/graphql/mutations/movie';
import { EDIT_REVIEW } from '~/graphql/mutations/movie';
import { MYREVIEWS_QUERY } from '~/graphql/queries/movies';

const props = defineProps({
  review: {
    type: Object,
    required: true
  },
  user_id: {
    type: Number,
  }
});
const emit = defineEmits(['close'])

const { mutate, onDone } = useMutation(EDIT_REVIEW)

onDone(() => {
  useToast().add({
    title: 'Review updated',
    description: 'Your review has been updated successfully',
    color: 'green'
  })
  closeModal()
})

const onEdit = () => {
  console.log('onEdit')
  const set = {
    rating: rating.value,
    feedback: feedback.value
  }
  mutate({
    id: props.review.id,
    set
  }, {
    refetchQueries: [{query: MYREVIEWS_QUERY, variables: {_eq: props.user_id}}]
  })
}

const closeModal = () => {
  emit('close')
}

const rating = ref(props.review.rating);
const feedback = ref(props.review.feedback);
</script>