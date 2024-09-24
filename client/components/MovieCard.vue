<template>
    <div class="group relative cursor-pointer overflow-hidden rounded-lg shadow-lg transition-transform duration-300 ease-in-out hover:scale-105 h-[300px]">
      <div 
        class="absolute inset-0 bg-contain bg-center h-full"
        :style="{ backgroundImage: `url(${props.movie.thumbnail})` }"
      >
        <div class="absolute inset-0 bg-black-950 bg-opacity-40 transition-opacity duration-300 group-hover:bg-opacity-30"></div>
      </div>
      <div class="relative flex h-full flex-col justify-end p-2">
        <div class="space-y-2 relative px-1 group-hover:backdrop-blur-lg group-hover:bg-black-950 group-hover:bg-opacity-40">
          <h3 class="text-2xl font-bold text-white px-3">{{ props.movie.title }}</h3>
          <div class="flex items-center justify-between space-x-2 text-sm text-gray-200">
            <span class="flex items-center">
              <CalendarIcon class="w-4 h-4 pr-1"/>
              <p v-if="props.type === 'home'">{{ formatDateShort((props.movie as MovieHome).scheduleDate) }}</p>
              <p v-else>Release: {{ formatYear(props.movie.releaseDate) }}</p>
            </span>
            <span class="flex items-center">
              <TimerIcon class="w-4 h-4 pr-1"/>
              {{ props.movie.duration }}
            </span>
            <span v-if="props.type === 'home'" class="inline-flex items-center rounded-lg bg-yellow px-2.5 py-0.5 text-xs font-medium">
              {{ (props.movie as MovieHome).price }} Birr
            </span>
            <span v-else class="inline-flex items-center rounded-lg bg-yellow-500 px-2.5 py-0.5 text-xs font-medium text-black-950">
              {{ props.movie.genre }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </template>
  
  <script setup lang="ts">
  import { CalendarIcon, TimerIcon } from 'lucide-vue-next';
  import type { Movie, MovieHome } from '~/types';
  
  const props = defineProps({
    movie: {
      type: Object as () => MovieHome | Movie,
      required: true
    },
    type: {
      type: String,
      dfault: 'movie'
    },
  });
</script>