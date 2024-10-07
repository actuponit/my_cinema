<template>
  <UModal :ui="{ divide: 'divide-y divide-gray-100 dark:divide-gray-800' }" class="w-[50%]">
    <UCard>
      <template #header>
        <div class="flex justify-between items-center">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
            Add Schedule
          </h3>
          <UButton color="gray" variant="ghost" icon="i-heroicons-x-mark-20-solid" @click="emit('close')" />
        </div>
      </template>
      <TestScheduleForm name="schedules" v-model="schedules" @update:model-value="schedules = $event" rows />
      <template #footer>
        <div class="flex justify-end space-x-3">
          <UButton color="gray" variant="outline" @click="emit('close')">
           Cancel
          </UButton>
          <UButton :loading="loading" color="primary" @click="addSchedules">
            Add schedules
          </UButton>
        </div> 
      </template>
    </UCard>
  </UModal>
</template>
<script setup lang="ts">

const props = defineProps({
  id: String
})

const schedules = ref([])

const emit = defineEmits(['close'])

const {executeInsert, onDone, loading} = useAddSchedule();

onDone(() => {
  useToast().add({
    title: "Opperation succesfull!",
    description: "Add all the cinema schedules for this movie",
    color: "green"
  })
  emit('close')

})

const addSchedules = async () => {
  let schedulesData = schedules.value.filter((schedule: any) => {
    return !(!schedule.price || schedule.price < 10 || schedule.hall === '' || schedule.format === '' || schedule.date === '') })
    .map((schedule: any) => {
    return {
      start_time: schedule.date,
      hall: schedule.hall,
      movie: props.id as string,
      format: cinemaFormat(schedule.format),
      price: schedule.price
    }
  })
  await executeInsert(schedulesData)
  console.log(schedulesData)
}
</script>