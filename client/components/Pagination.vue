<template>
    <div class="flex justify-center items-center space-x-4">
        <button
          @click="emitPrevPage"
          :disabled="currentPage === 1"
          class="px-3 py-1 bg-primary text-white rounded-md disabled:opacity-50 hover:bg-opacity-80 transition-colors duration-200"
        >
          Previous
        </button>
        <span class="text-lg">Page {{ currentPage }} of {{ totalPages }}</span>
        <button
          @click="emitNextPage"
          :disabled="currentPage === totalPages"
          class="px-3 py-1 bg-primary text-white rounded-md disabled:opacity-50 hover:bg-opacity-80 transition-colors duration-200"
        >
          Next
        </button>
    </div>
</template>

<script setup>
const props = defineProps({
    currentPage: Number,
    totalPages: Number,
});
const emit = defineEmits(['prevPage', 'nextPage']);

const emitPrevPage = () => emit('prevPage');
const emitNextPage = () => emit('nextPage');
</script>