<template>
    <UTooltip :text="ticket.id" :ui="{background: 'blue'}">
      <div class="relative w-full max-w-md mx-auto overflow-hidden rounded-xl shadow-lg">
        <!-- Blurred background -->
        <div 
          class="absolute inset-0 bg-cover bg-center filter blur-sm"
          :style="{ backgroundImage: `url(${ticket.thumbnail})` }"
        ></div>
        
        <!-- Gradient overlay -->
        <div class="absolute inset-0 bg-gradient-to-br from-gray-800/80 via-black-950/80 to-blue-900/80"></div>
        
        <!-- Content -->
        <div class="relative p-6 text-white">
          <div class="flex justify-between items-start mb-4">
            <h2 class="text-3xl font-bold leading-tight">{{ ticket.movieTitle }}</h2>
            <div>
              <span v-if="ticket.isUpcoming"
                class="px-2 py-1 text-xs mr-3 font-semibold rounded-full bg-primary text-white"
              >
                Upcoming
              </span>
            </div>
          </div>
          
          <div class="flex items-center mb-4">
            <ClockIcon class="w-5 h-5 mr-2" />
            <span>{{ formatDateFull(ticket.movieTime) }}</span>
          </div>
          
          <div class="flex items-center mb-4">
            <Popcorn class="w-5 h-5 mr-2" />
            <span>Format: {{ ticket.format }}</span>
          </div>
          
          <div class="flex justify-between items-center mb-6">
            <div class="flex items-center">
              <MapPinIcon class="w-5 h-5 mr-2" />
              <span>{{ ticket.hall }}</span>
            </div>
            <div class="text-right">
              <div class="text-2xl font-bold">${{ ticket.price * ticket.amount }}</div>
              <div class="text-sm opacity-75">{{ ticket.amount }} ticket(s)</div>
            </div>
          </div>
          
          <div class="flex items-center p-4 bg-white/10 rounded-lg backdrop-blur-sm">
            <div class="w-1/2">
              <div class="text-sm opacity-75">Ticket ID</div>
              <div class="font-mono truncate">{{ ticket.id }}</div>
            </div>
            <div class="text-right flex-1">
              <div class="text-sm opacity-75">Purchased</div>
              <div>{{ formatDateFull(ticket.boughtAt) }}</div>
            </div>
          </div>
        </div>
        
        <!-- Decorative elements -->
        <div class="absolute top-0 right-0 w-20 h-20 bg-white/20 rounded-full -translate-y-1/2 translate-x-1/2"></div>
        <div class="absolute bottom-0 left-0 w-16 h-16 bg-white/20 rounded-full translate-y-1/2 -translate-x-1/2"></div>
      </div>
    </UTooltip>
  </template>
  
  <script setup lang="ts">
  import { ref } from 'vue'
  import { ClockIcon, MapPinIcon, Move3dIcon, Popcorn } from 'lucide-vue-next'
  import type { Ticket } from '~/types';
	import { formatDateFull } from '~/utils/dateformat';
  
  const props = defineProps({
    ticket: {
      type: Object as ()=> Ticket,
      required: true
    }
  })

	const deleteTicket = () => {
		console.log('Deleting ticket', ticket.value.id);
	}

	const ticket = ref(props.ticket)
</script>