<template>
    <div class="w-full flex-1 min-h-full p-8">
        <div class="flex justify-between items-center">
            <h1 class="text-3xl font-bold mb-5">Tickets</h1>
            <UButton @click="goBack" variant="link">Back</UButton>
        </div>
        <div class="flex gap-6">
            <h1 class="text-xl font-bold mb-5 text-gray-400">{{ "Movie: Inception" }}</h1>
            <h1 class="text-xl font-bold mb-5 text-gray-400">{{ "  Schedule: @Sep/22/2024 Hall A 2D " }}</h1>

        </div>
        <UTable class="w-full" :rows="tickets" :columns="columns"/>
    </div>
</template>
<script setup lang="ts">
const tickets = [
    {id: "txtaafjdks", fullname: "Imran Mohammed", email: "imo@gmail", bookedAt: "2020/3232/323", quantity: 4},
]
const columns = [
    {key: "id", label: "ID"},
    {key: "fullname", label: "Fullname"},
    {key: "email", label: "Email"},
    {key: "bookedAt", label: "Booked At"},
    {key: "quantity", label: "Quantity"},
]
const router = useRouter()
const goBack = () => router.back()
</script>