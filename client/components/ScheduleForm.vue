<template>
  <div class="flex-1 w-full text-center col-span-2">
    <div v-for="(schedule, index) in schedules" :key="index" class="mb-4 p-4 border border-gray-200 rounded-md">
      <div class="grid grid-cols-1 gap-6 sm:grid-cols-3">
        <UFormGroup label="Date" :name="`date-${index}`">
          <UInput :id="`date-${index}`" v-model="schedule.date" type="date" />
        </UFormGroup>
        <UFormGroup label="Cinema Hall" :name="`hall-${index}`">
          <USelect :id="`hall-${index}`" v-model="schedule.hall" :options="cinemaHalls" />
        </UFormGroup>
        <UFormGroup label="Price" :name="`price-${index}`">
          <UInput :id="`price-${index}`" v-model.number="schedule.price" type="number" step="0.1" />
        </UFormGroup>
      </div>
      <div class="w-full ml-auto text-end">
        <UButton @click="removeSchedule(index)" variant="ghost" color="red" class="mt-2 text-sm">Remove</UButton>
      </div>
    </div>
    <UButton @click="addSchedule" icon="i-heroicons-plus">
      Add Schedule
    </UButton>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'

const props = defineProps({
  modelValue: {
    type: Array,
    default: () => []
  },
  name: String
})

const emit = defineEmits(['update:modelValue'])

const schedules = ref(props.modelValue)

watch(schedules, (newValue) => {
  emit('update:modelValue', newValue)
}, { deep: true })


const cinemaHalls = ['Hall A', 'Hall B', 'Hall C', 'Hall D']

const addSchedule = () => {
  schedules.value.push({ date: '', hall: '', price: null })
}

const removeSchedule = (index) => {
  schedules.value.splice(index, 1)
}
</script>