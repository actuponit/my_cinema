<template>
    <div class="flex-1 w-full text-center col-span-2">
      <div v-for="(schedule, index) in fields" :key="schedule.key" class="mb-4 p-4 border border-gray-200 rounded-md">
        <div class="grid grid-cols-1 gap-6 sm:grid-cols-3">
          <UFormGroup label="Schedule Date and time" :name="`date-${index}`" :error="displayError('date', index)">
            <VueDatePicker v-model="schedule.value.date" :is24="false" :min-date="new Date()" dark/>
          </UFormGroup>
          <UFormGroup label="Cinema Hall" :name="`hall-${index}`" :error="displayError('hall', index)">
            <USelectMenu :name="`hall-${index}`" :id="`hall-${index}`" v-model="schedule.value.hall" :options="cinemaHalls" />
          </UFormGroup>
          <UFormGroup label="Price" :name="`price-${index}`" :error="displayError('price', index)">
            <UInput :name="`price-${index}`" :id="`price-${index}`" v-model="schedule.value.price" type="number" step="0.1" />
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
  
<script setup lang="ts">
  import { useFieldArray, useForm } from 'vee-validate';
  import * as yup from 'yup';
  import VueDatePicker from '@vuepic/vue-datepicker';
  import '@vuepic/vue-datepicker/dist/main.css';

  const props = defineProps({
    modelValue: {
      type: Array,
      default: () => []
    },
    name: {
      type: String,
      required: true
    }
  });

  const cinemaHalls = ['Hall A', 'Hall B', 'Hall C', 'Hall D']

  const schema = yup.object({
    schedules: yup.array().of(
      yup.object({
        date: yup.date()
          .typeError('Invalid date format')
          .min(new Date(), 'Date cannot be in the past')
          .required('Date is required'),
        hall: yup.string().required('Cinema hall is required').oneOf(cinemaHalls, 'Invalid cinema hall'),
        price: yup.number()
        .typeError('Must be a number')
        .required('Price is required').min(10, 'Price must be greater than 0')
      })
    )
  });

  const { errors, meta } = useForm({
    validationSchema: schema,
    initialValues: {
      schedules: [{ date: '', hall: '', price: null }]
    }
  });

  interface Schedule {
    date: string;
    hall: string;
    price: number;
  }

  const { fields, push, remove } = useFieldArray<Schedule>(() => props.name)

  const emit = defineEmits(['update:modelValue'])
  
  watch(fields, (newValue) => {
    emit('update:modelValue', newValue.map((schedule) => schedule.value))
  }, { deep: true })
  
  const addSchedule = () => {
    console.log('addSchedule')
    push({ date: '', hall: '', price: 0 })
    console.log("fields", fields.value)
  }
  
  const removeSchedule = (index: number) => {
    remove(index)
  }
  const displayError = (field: string, index: number) => {
    console.log('displayError', meta.value)
    if (meta.value.dirty) {
      return errors.value[`schedules[${index}].${field}`] || ''
    }
    return ''
  } 
    
</script>