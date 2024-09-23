<template>
  <UModal>
    <UCard :ui="{ divide: 'divide-y divide-gray-100 dark:divide-gray-800' }">
      <template #header>
        <div class="flex justify-between items-center">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
            Book Tickets
          </h3>
          <UButton color="gray" variant="ghost" icon="i-heroicons-x-mark-20-solid" @click="closeModal" />
        </div>
      </template>

      <div class="p-4 space-y-4">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <p class="text-sm font-medium text-gray-500 dark:text-gray-400">Date</p>
            <p class="mt-1 text-sm text-gray-900 dark:text-white">{{ movieDetails.date }}</p>
          </div>
          <div>
            <p class="text-sm font-medium text-gray-500 dark:text-gray-400">Time</p>
            <p class="mt-1 text-sm text-gray-900 dark:text-white">{{ movieDetails.time }}</p>
          </div>
          <div>
            <p class="text-sm font-medium text-gray-500 dark:text-gray-400">Format</p>
            <UBadge :color="'primary'" class="mt-1">
              {{ movieDetails.format }}
            </UBadge>
          </div>
          <div>
            <p class="text-sm font-medium text-gray-500 dark:text-gray-400">Hall</p>
            <p class="mt-1 text-sm text-gray-900 dark:text-white">{{ movieDetails.hall }}</p>
          </div>
        </div>

        <div>
          <p class="text-sm font-medium text-gray-500 dark:text-gray-400">Price per ticket</p>
          <p class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">${{ movieDetails.price }}</p>
        </div>

        <UFormGroup label="Number of Tickets" name="ticketCount" description="Maximum 5 tickets per booking">
          <UInput
            v-model="ticketCount"
            type="number"
            min="1"
            max="5"
            placeholder="Enter number of tickets"
          />
        </UFormGroup>

        <div v-if="ticketCount > 0" class="bg-gray-50 dark:bg-gray-800 p-3 rounded-md">
          <p class="text-sm font-medium text-gray-700 dark:text-gray-300">Total Price</p>
          <p class="text-lg font-bold text-gray-900 dark:text-white">
            ${{ (movieDetails.price * ticketCount).toFixed(2) }}
          </p>
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end space-x-3">
          <UButton color="gray" variant="outline" @click="closeModal">
            Cancel
          </UButton>
          <UButton color="primary" :disabled="!isValid">
            Book Tickets
          </UButton>
        </div> 
      </template>
    </UCard>
  </UModal>
</template>

<script setup lang="ts">
  import { ref, computed } from 'vue'
  import type { Schedule } from '~/types'

  const props = defineProps({
    movieDetails: {
      type: Object as () => Schedule,
      required: true
    }
  })

  const emit = defineEmits(['close'])

  const ticketCount = ref(1)

  const isValid = computed(() => ticketCount.value > 0 && ticketCount.value <= 5)

  const closeModal = () => {
    emit('close')
  }

  const bookTickets = () => {
    if (isValid.value) {
      closeModal()
    }
  }
</script>