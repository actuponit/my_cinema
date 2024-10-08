<template>
    <div class="space-y-2 max-w-fit">
        <p class="text-sm font-medium text-gray-300">Choose price range</p>
        <div class="flex items-center bg-gray-800 rounded-md">
            <input
            v-model="minPrice"
            @change="updateMinPrice"
            class="bg-gray-800 text-white w-28 px-4 py-2 border-blue-500 border border-solid rounded-l-md focus:outline-none focus:ring-2 focus:ring-primary"
            placeholder="200"
            type="number"
            />
            <span class="text-gray-400 px-2">to</span>
            <input
            v-model="maxPrice"
            @change="updateMaxPrice"
            class="bg-gray-800 text-white px-4 py-2 rounded-r-md border-blue-500 border border-solid focus:outline-none focus:ring-2 focus:ring-primary w-28"
            placeholder="300"
            type="number"
            />
        </div>
    </div>
</template>
<script setup lang="ts">
const props = defineProps({
    minPrice: {
        type: Number,
    },
    maxPrice: {
        type: Number,
    }
});
const minPrice = ref(props.minPrice);
const maxPrice = ref(props.maxPrice);
const emit = defineEmits(['update:minPrice', 'update:maxPrice']);
const updateMaxPrice = () => {
    emit('update:maxPrice', maxPrice.value);
};
const updateMinPrice = () => {
    emit('update:maxPrice', maxPrice.value);
};
</script>