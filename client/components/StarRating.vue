<template>
    <div class="flex items-center">
      <span class="mr-3 text-white" v-if="showRating">
        {{ (hoverRating || rating || 0).toFixed(1) }}
      </span>
      <div 
        v-for="star in 5" 
        :key="star" 
        class="relative cursor-pointer"
        @click="setRating(star)"
        @mouseover="setHover(star)"
        @mouseleave="hoverRating = 0"
      >
        <!-- Unfilled star -->
        <svg
          class="w-4 h-4 text-yellow transition-colors duration-200"
          :class="{ 'text-primary': (hoverRating || rating) >= star }"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z"
          ></path>
        </svg>
        <!-- Filled star -->
        <div
          class="absolute top-0 left-0 overflow-hidden transition-all duration-200 text-yellow text-"
          :style="{ width: `${getFillPercentage(star)}%` }"
        >
          <svg
            class="w-4 h-4"
            fill="currentColor"
            viewBox="0 0 24 24"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z"
            ></path>
          </svg>
        </div>
      </div>
    </div>
</template>

<script setup>
  import { ref, computed } from 'vue';
  
  const props = defineProps({
    initialRating: {
      type: Number,
      default: 0,
      validator: (value) => value >= 0 && value <= 5
    },
    showRating: {
      type: Boolean,
      default: true
    },
    readOnly: {
      type: Boolean,
      default: false
    }
  });
  
  const emit = defineEmits(['update:rating']);
  
  const rating = ref(props.initialRating);
  const hoverRating = ref(0);
  
  const getFillPercentage = computed(() => (starIndex) => {
    const currentRating = hoverRating.value || rating.value;
    const fillPercentage = (currentRating - (starIndex - 1)) * 100;
    return Math.max(0, Math.min(100, fillPercentage));
  });
  
  const setHover = (newHoverRating) => {
    if (props.readOnly) return;
    hoverRating.value = newHoverRating;
  };
  const setRating = (newRating) => {
    if (props.readOnly) return;
    rating.value = newRating;
    emit('update:rating', newRating);
  };
</script>