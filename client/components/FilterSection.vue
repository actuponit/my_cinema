<template>
    <div class="bg-gray-900 p-3 rounded-lg shadow-xl">
        <div class="flex justify-between">
            <h2 class="text-lg font-bold text-white mb-3">Find schedules</h2>
            <button class="text-white bg-blue px-4 py-2 rounded-md">Apply</button>
        </div>
        <div class="flex justify-between">
            <div class="w-72 space-y-2">
                <label class="text-sm font-medium text-gray-300 block">Search by title</label>
                <input 
                    class="bg-gray-800 text-white px-4 py-2 rounded-md h-10 w-full focus:outline-none focus:ring-2 focus:ring-primary self-center"
                    placeholder="Inception"
                />
            </div>
            <PriceRange />
            <DateRange />
        </div>
    </div>
</template>
<script setup>
</script>