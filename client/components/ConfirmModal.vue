<template>
    <UModal class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
        <div class="p-6 rounded-lg shadow-xl">
            <h3 class="text-xl font-bold mb-4">{{ title }}</h3>
            <p class="mb-4">{{ message }}</p>
            <div class="flex justify-end space-x-2">
            <UButton @click="onAction" color="red" variant="outline">
                {{ action }}
            </UButton>
            <UButton @click="cancelDelete" color="gray">
                Cancel
            </UButton>
            </div>
        </div>
    </UModal>
</template>
<script setup>
const props = defineProps({
    message: String,
    title: String,
    onAction: Function,
    action: String
})
const emit = defineEmits(['close'])

const cancelDelete = () => {
    emit('close')
}
</script>