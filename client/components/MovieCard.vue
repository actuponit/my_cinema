<template>
    <div class="group relative cursor-pointer overflow-hidden rounded-lg shadow-lg transition-transform duration-300 ease-in-out hover:scale-105 h-[300px]">
      <div 
        class="absolute inset-0 bg-contain bg-center h-full"
        :style="{ backgroundImage: `url(${thumbnailUrl})` }"
      >
        <div class="absolute inset-0 bg-black-950 bg-opacity-40 transition-opacity duration-300 group-hover:bg-opacity-30"></div>
      </div>
      <div class="relative flex h-full flex-col justify-end p-2">
        <div class="space-y-2 relative px-1 group-hover:backdrop-blur-lg group-hover:bg-black-950 group-hover:bg-opacity-40">
          <h3 class="text-2xl font-bold text-white px-3">{{ title }}</h3>
          <div class="flex items-center justify-between space-x-2 text-sm text-gray-200">
            <span class="flex items-center">
              <CalendarIcon class="w-4 h-4 pr-1"/>
              <time :datetime="date">{{ formattedDate }}</time>
            </span>
            <span class="flex items-center">
              <TimerIcon class="w-4 h-4 pr-1"/>
              3h 30m
            </span>
            <span class="inline-flex items-center rounded-lg bg-yellow px-2.5 py-0.5 text-xs font-medium text-black">
              {{ "250 Birr" }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </template>
  
  <script setup>
  import { computed } from 'vue';
  import { CalendarIcon, TimerIcon } from 'lucide-vue-next';
  
  const props = defineProps({
    title: {
      type: String,
      required: true
    },
    date: {
      type: String,
      required: true
    },
    category: {
      type: String,
      required: true
    },
    thumbnailUrl: {
      type: String,
      required: true
    }
  });
  
  const formattedDate = computed(() => {
    return new Date(props.date).toLocaleDateString('en-US', { 
      year: 'numeric', 
      month: 'short', 
      day: 'numeric' 
    });
  });
  </script>