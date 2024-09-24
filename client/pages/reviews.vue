<template>
    <div class="min-h-screen bg-gray-900 text-gray-100 flex-1">
      <div class="container mx-auto px-4 py-8">
        <h1 class="text-4xl font-bold mb-8 text-center">Your Reviews</h1>
        
        <div v-if="reviews.length > 0">
          <div class="flex justify-between items-center mb-6">
            <USelect
              v-model="sortBy"
              :options="sortOptions"
              placeholder="Sort by"
              class="w-40"
            />
            <UPagination
              v-model="currentPage"
              :total="totalPages"
              :ui="{ rounded: 'rounded-full' }"
            />
          </div>
          
          <TransitionGroup
            name="list"
            tag="ul"
            class="grid gap-6 lg:grid-cols-3"
          >
            <li
              v-for="review in reviews"
              :key="review.id"
              class="bg-gray-800 rounded-lg p-6 shadow-lg transition-all duration-300 hover:shadow-xl hover:scale-105"
            >
              <div class="flex justify-between items-start mb-4">
                <h3 class="text-xl font-semibold">{{ review.productName }}</h3>
                <UBadge color="yellow" size="lg">
                  {{ review.rating }}/5
                </UBadge>
                
              </div>
              <p class="text-gray-300 mb-4">{{ review.content }}</p>
              <div class="flex justify-between items-center text-sm text-gray-400">
                <span>{{ formatDateFull(review.date) }}</span>
                <UButton
                  color="gray"
                  variant="ghost"
                  icon="i-heroicons-pencil-square"
                  :ui="{ rounded: 'rounded-full' }"
                >
                  Edit
                </UButton>
              </div>
            </li>
          </TransitionGroup>
          
          <div class="mt-8 flex justify-center">
            <UPagination
              v-model="currentPage"
              :total="totalPages"
              :ui="{ rounded: 'rounded-full' }"
            />
          </div>
        </div>
        <ErrorComponent v-else icon="i-heroicons-chat-bubble-bottom-center-text" title="No reviews yet"/>
      </div>
    </div>
  </template>
  
<script setup lang="ts">
  import { ref, computed } from 'vue'
  import { formatDateFull } from '~/utils/dateformat'
  
  // Mock data for reviews
  const reviews = ref([
    { id: 1, productName: 'Wireless Earbuds', rating: 4, content: 'Great sound quality and comfortable fit.', date: '2023-09-15' },
    { id: 2, productName: 'Smart Watch', rating: 5, content: 'Amazing features and battery life!', date: '2023-08-22' },
    { id: 3, productName: 'Laptop Stand', rating: 3, content: 'Decent build quality, but could be more adjustable. Decent build quality, but could be more adjustable. Decent build quality, but could be more adjustable. Decent build quality, but could be more adjustable.', date: '2023-07-30' },
    // Add more mock reviews here...
  ])

  const sortOptions = [
    { label: 'Newest', value: 'date' },
    { label: 'Highest Rated', value: 'rating' },
    { label: 'Product Name', value: 'productName' },
  ]
  
  const sortBy = ref('date')
  const currentPage = ref(1)
  const itemsPerPage = 6
  
  const totalPages = computed(() => Math.ceil(reviews.value.length / itemsPerPage))
</script>
  
<style scoped>
  .list-enter-active,
  .list-leave-active {
    transition: all 0.5s ease;
  }
  .list-enter-from,
  .list-leave-to {
    opacity: 0;
    transform: translateY(30px);
  }
</style>