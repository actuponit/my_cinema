<template>
    <div class="bg-gray-800 p-3 rounded-lg shadow-xl">
        <div class="flex justify-between">
            <h2 class="text-lg font-bold text-white mb-3">Find schedules</h2>
            <button class="text-white bg-primary px-4 py-2 rounded-md">Apply</button>
        </div>
        <div class="lg:flex justify-between">
            <div class="w-72 space-y-2 relative">
                <label class="text-sm font-medium text-gray-300 block">Search by title</label>
                <input
                    class="bg-gray-800 text-white border-blue-500 border px-4 py-2 rounded-md h-10 w-full focus:outline-none focus:ring-2 focus:ring-primary self-center"
                    placeholder="Inception"
                />
                <SearchIcon class="absolute right-2 bottom-3 w-5 h-5 text-primary"/>
            </div>
            <PriceRange />
            <div class="space-y-2">
                <label class="text-sm font-medium text-gray-300 block">Choose date range</label>
                <VueDatePicker v-model="selected" :enable-time-picker="false" :min-date="new Date()" range dark/>
            </div>
        </div>
    </div>
</template>
<script setup lang="ts">
    import VueDatePicker from '@vuepic/vue-datepicker';
    import '@vuepic/vue-datepicker/dist/main.css';
import { SearchIcon } from 'lucide-vue-next';
    const selected = ref();
</script>

<style>
.dp__theme_dark {
    --dp-primary-color: rgb(var(--color-primary-500));
    --dp-border-color-focus: rgb(var(--color-primary-500));
    --dp-border-color: rgb(var(--color-primary-500));
    --dp-background-color: rgb(var(--color-gray-800));
}
</style>
