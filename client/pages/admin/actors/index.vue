<template>
    <div class="w-full flex-1 min-h-full p-8">
        <div class="flex justify-between items-center">
            <h1 class="text-3xl font-bold mb-5">Casts And Crews</h1>
        </div>
        <UTable class="w-full" :rows="tickets" :columns="columns" >
            <template #actions-data="{ row }">
                <div class="flex space-x-2 justify-around">
                    <UButton variant="link">
                        Edit Info
                    </UButton>
                    <button @click="" class="px-3 py-1 bg-red-600 text-white rounded-md hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2 focus:ring-offset-gray-800">
                        Delete
                    </button>
                </div>
            </template>
        </UTable>
    </div>
</template>
<script setup lang="ts">
const tickets = [
    {id: "txtaafjdks", firstname: "Imran", lastname: "Mohammed", movies: 4, role: "Actor"},
    {id: "txtaafjdks", firstname: "Imran", lastname: "Mohammed", movies: 4, role: "Director"},
]
const columns = [
    {key: "firstname", label: "First Name"},
    {key: "lastname", label: "Last Name"},
    {key: "role", label: "Role"},
    {key: "movies", label: "Number of Movies"},
    {key: "actions", label: ""},
]
</script>