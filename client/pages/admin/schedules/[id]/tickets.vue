<template>
    <div class="w-full flex-1 min-h-full p-8">
        <div class="flex justify-between items-center">
            <h1 class="text-3xl font-bold mb-5">Tickets</h1>
            <UButton @click="goBack" variant="link">Back</UButton>
        </div>
        <div class="flex gap-6">
            <h1 class="text-xl font-bold mb-5 text-gray-400">{{ `Movie: ${schedule.movie}` }}</h1>
            <h1 class="text-xl font-bold mb-5 text-gray-400">{{ ` Schedule: @${schedule.time}, Venue ${schedule.hall}, Format ${schedule.format}` }}</h1>
        </div>
        <UTable class="w-full" :loading="loading" :rows="mytickets" :columns="columns"/>
    </div>
</template>
<script setup lang="ts">
import { SCHEDULES_QURY_BYID } from '~/graphql/queries/schedule';
const id = useRoute().params.id;
const { result, loading } = useQuery(SCHEDULES_QURY_BYID, {id: id});

console.log("result", result.value)
const mytickets = computed(() => {
    return result.value?.schedules_by_pk.tickets.map((t: any) => ({
        id: t.id,
        fullname: t.user.first_name + " " + t.user.last_name,
        email: t.user.email,
        bookedAt: t.created_at,
        quantity: t.quantity,
    })) || []
})

const schedule = computed(() => {
    return {
        movie: result.value.schedules_by_pk.movieByMovie.title,
        time: formatDateFull(result.value?.schedules_by_pk.start_time),
        hall: result.value?.schedules_by_pk.hall,
        format: cinemaFormatReverse(result.value?.schedules_by_pk.format),
        price: result.value?.schedules_by_pk.price
    }
})

console.log(schedule.value)
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