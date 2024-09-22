<template>
    <div class="space-y-2">
        <label for="date-range" class="block text-sm font-medium text-gray-300">
            Select a date range
        </label>
        <DatePicker
          v-model="range"
          :min-date="new Date()"
          :max-date="maxDate"
          :masks="masks"
          is-range
          is-dark
          :color="blue"
          trim-weeks
          class="date-range-picker"
        >
          <template v-slot="{ inputValue, inputEvents }">
            <div class="flex items-center bg-gray-800 rounded-md">
              <input
                id="date-range"
                :value="inputValue.start"
                v-on="inputEvents.start"
                class="bg-gray-800 text-white px-4 py-2 border-black border-solid border-2 rounded-l-md focus:outline-none focus:ring-2 focus:ring-primary"
                placeholder="Start date"
                readonly
              />
              <span class="text-gray-400 px-2">to</span>
              <input
                :value="inputValue.end"
                v-on="inputEvents.end"
                class="bg-gray-800 text-white px-4 py-2 rounded-r-md border-black border-solid border-2 focus:outline-none focus:ring-2 focus:ring-primary"
                placeholder="End date"
                readonly
              />
            </div>
          </template>
        </DatePicker>
    </div>
</template>
  
<script setup>
  import { ref, computed } from 'vue';
  import { DatePicker } from 'v-calendar';
  import 'v-calendar/dist/style.css';
  
  const range = ref({ start: null, end: null });
  
  const maxDate = computed(() => {
    const date = new Date();
    date.setFullYear(date.getFullYear() + 1);
    return date;
  });
  
  const masks = {
    input: 'YYYY-MM-DD',
  };
  
  const formatDate = (date) => {
    if (!date) return '';
    return new Date(date).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };
  console.log(range.value)
</script>
<style>
:root {
  --vc-gray-50: rgb(var(--color-gray-50));
  --vc-gray-100: rgb(var(--color-gray-100));
  --vc-gray-200: rgb(var(--color-gray-200));
  --vc-gray-300: rgb(var(--color-gray-300));
  --vc-gray-400: rgb(var(--color-gray-400));
  --vc-gray-500: rgb(var(--color-gray-500));
  --vc-gray-600: rgb(var(--color-gray-600));
  --vc-gray-700: rgb(var(--color-gray-700));
  --vc-gray-800: rgb(var(--color-gray-800));
  --vc-gray-900: rgb(var(--color-gray-900));
}

.vc-primary {
  --vc-accent-50: rgb(var(--color-primary-50));
  --vc-accent-100: rgb(var(--color-primary-100));
  --vc-accent-200: rgb(var(--color-primary-200));
  --vc-accent-300: rgb(var(--color-primary-300));
  --vc-accent-400: rgb(var(--color-primary-400));
  --vc-accent-500: rgb(var(--color-primary-500));
  --vc-accent-600: rgb(var(--color-primary-600));
  --vc-accent-700: rgb(var(--color-primary-700));
  --vc-accent-800: rgb(var(--color-primary-800));
  --vc-accent-900: rgb(var(--color-primary-900));
}
</style>
