<template>
  <div class="min-h-screen flex-1">
    <UContainer class="py-8">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white mb-8">My Booked Tickets</h1>

      <div v-if="tickets.length > 0" class="grid lg:grid-cols-3 gap-4">
        <CinemaTickets v-for="ticket in tickets" :ticket="ticket" :key="ticket.id" />
      </div>
      <ErrorComponent v-else icon="i-heroicons-archive-box-x-mark" title="You don't have any tickets booked" />
    </UContainer>

  </div>
</template>

<script setup lang="ts">
import { TICKET_QUERY } from '~/graphql/queries/tickets';

const { result, loading } = useQuery(TICKET_QUERY)
const tickets = computed(() => {
  console.log('tickets', result.value);
  return result.value?.tickets.map((ticket: any) => {
    return {
      movieTitle: ticket?.schedule?.movieByMovie?.title,
      hall: ticket?.schedule?.hall,
      thumbnail: displayImage(ticket?.schedule?.movieByMovie?.featured_image),
      price: ticket?.schedule?.price,
      amount: ticket?.quantity,
      id: ticket?.id,
      boughtAt: ticket?.created_at,
      movieTime: ticket?.schedule?.start_time,
      isUpcoming: new Date(ticket?.schedule?.start_time) > new Date(),
      format: cinemaFormatReverse(ticket?.schedule?.format)
    }
  }) || []
})
definePageMeta({
  middleware: ['must-login'],
});
</script>