<template>
    <div class="flex flex-col items-center justify-center p-8 text-center">
      <UIcon
        :name="icon"
        class="w-16 h-16 mb-4 text-gray-400 dark:text-gray-600"
      />
      <h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-2">
        {{ title }}
      </h2>
      <p class="text-gray-600 dark:text-gray-400 max-w-md">
        {{ message }}
      </p>
      <slot name="action"></slot>
    </div>
  </template>
  
  <script setup>
  defineProps({
    icon: {
      type: String,
      default: 'i-heroicons-exclamation-circle',
    },
    title: {
      type: String,
      required: true,
    },
    message: {
      type: String,
      required: true,
    },
  })
  </script>