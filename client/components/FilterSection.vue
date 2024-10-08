<template>
    <div class="bg-gray-800 p-3 rounded-lg shadow-xl">
        <div class="flex justify-between">
            <h2 class="text-lg font-bold text-white mb-3">Find schedules</h2>
            <button class="text-white bg-primary px-4 py-2 rounded-md" @click="onApplyFilters">Apply</button>
        </div>
        <div class="lg:flex justify-between">
            <div class="w-72 space-y-2 relative">
                <label class="text-sm font-medium text-gray-300 block">Search by title</label>
                <input
                    v-model="title"
                    class="bg-gray-800 text-white border-blue-500 border px-4 py-2 rounded-md h-10 w-full focus:outline-none focus:ring-2 focus:ring-primary self-center"
                    placeholder="Inception" />
                <SearchIcon class="absolute right-2 bottom-3 w-5 h-5 text-primary" />
            </div>
            <div class="space-y-2 max-w-fit">
                <p class="text-sm font-medium text-gray-300">Choose price range</p>
                <div class="flex items-center bg-gray-800 rounded-md">
                    <input v-model="minPrice"
                        class="bg-gray-800 text-white w-28 px-4 py-2 border-blue-500 border border-solid rounded-l-md focus:outline-none focus:ring-2 focus:ring-primary"
                        placeholder="200" type="number" />
                    <span class="text-gray-400 px-2">to</span>
                    <input v-model="maxPrice"
                        class="bg-gray-800 text-white px-4 py-2 rounded-r-md border-blue-500 border border-solid focus:outline-none focus:ring-2 focus:ring-primary w-28"
                        placeholder="300" type="number" />
                </div>
            </div>
            <div class="space-y-2">
                <label class="text-sm font-medium text-gray-300 block">Choose date range</label>
                <VueDatePicker v-model="selected" :enable-time-picker="false" :min-date="new Date()" range dark />
            </div>
        </div>
    </div>
</template>
<script setup lang="ts">
import VueDatePicker from '@vuepic/vue-datepicker';
import '@vuepic/vue-datepicker/dist/main.css';
import { SearchIcon } from 'lucide-vue-next';

const selected = ref();
const minPrice = ref();
const maxPrice = ref();
const title = ref('');
const filters = reactive(
    { selected, minPrice, maxPrice, title }
);

const emit = defineEmits(['update:filters'])
const onApplyFilters = () => {
    console.log("title ", title.value)
    emit('update:filters', filters)
}

</script>

<style>
.dp__theme_dark {
    --dp-primary-color: rgb(var(--color-primary-500));
    --dp-border-color-focus: rgb(var(--color-primary-500));
    --dp-border-color: rgb(var(--color-primary-500));
    --dp-background-color: rgb(var(--color-gray-800));
}
</style>
