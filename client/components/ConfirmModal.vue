<template>
    <UModal class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
        <div class="p-6 rounded-lg shadow-xl">
            <h3 class="text-xl font-bold mb-4">Confirm Deletion</h3>
            <p class="mb-4">Are you sure you want to delete schedule {{ schedule.movie + ' @' +schedule.date+':'+schedule.time }}?</p>
            <div class="flex justify-end space-x-2">
            <UButton @click="confirmDelete" color="red" variant="outline">
                Delete
            </UButton>
            <UButton @click="cancelDelete" color="gray">
                Cancel
            </UButton>
            </div>
        </div>
    </UModal>
</template>
<script setup>
    defineProps({
        schedule: Object,
    })
    const emit = defineEmits(['close'])
    const confirmDelete = () => {
        emit('close')
        console.log('Deleting schedule', schedule)
        // Add logic to delete schedule
    }
    const cancelDelete = () => {
        emit('close')
    }
</script>