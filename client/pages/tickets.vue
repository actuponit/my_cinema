<template>
    <div class="min-h-screen bg-gray-100 dark:bg-gray-900 flex-1">
      <UContainer class="py-8">
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white mb-8">My Booked Tickets</h1>
  
        <div v-if="tickets.length > 0" class="grid lg:grid-cols-3 gap-4">
          <CinemaTickets v-for="ticket in tickets" :ticket="ticket" :key="ticket.id"/>
        </div>
        <ErrorComponent v-else icon="i-heroicons-archive-box-x-mark" title="You don't have any tickets booked"/>
      </UContainer>

    </div>
  </template>
  
<script setup lang="ts">
  import { ref } from 'vue'
  import type { Ticket } from '~/types';
  
  // Mock data for booked tickets
  const tickets = ref<Ticket[]>([{
      movieTitle: "Inception",
      hall: "A3",
      thumbnail: "placeholder.webp",
      price: 12.99,
      amount: 2,
      id: "TKT-12345-6789",
      boughtAt: new Date("2023-06-15T14:30:00"),
      movieTime: new Date("2023-06-20T19:00:00"),
      isUpcoming: true
    },
    {
      movieTitle: "Inception",
      hall: "A3",
      thumbnail: "https://example.com/inception-thumbnail.jpg",
      price: 12.99,
      amount: 2,
      id: "TKT-12345-6780",
      boughtAt: new Date("2023-06-15T14:30:00"),
      movieTime: new Date("2023-06-20T19:00:00"),
      isUpcoming: false
    }
  ])

</script>