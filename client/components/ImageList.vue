<template>
  <div>
    <div v-if="images.length > 0" class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
      <div v-for="(image, index) in images" :key="index" class="relative group">
        <img :src="displayImage(image)" alt="actor image" class="w-full h-40 object-cover rounded-lg shadow-md" />
        <div class="absolute inset-0 bg-black bg-opacity-50 opacity-0 group-hover:opacity-100 transition-opacity duration-200 flex items-center justify-center rounded-lg">
          <UButton
            color="red"
            variant="solid"
            icon="i-heroicons-trash"
            size="sm"
            @click="deleteImage(index)"
          >
            Delete
          </UButton>
        </div>
      </div>
    </div>
    <p v-else class="text-gray-500 dark:text-gray-400">No images available</p>
  </div>
</template>

<script setup>
const props = defineProps({
  initialImages: {
    type: Array,
    default: () => []
  }
})

const images = ref(props.initialImages)

const deleteImage = (index) => {
  images.value.splice(index, 1)
}
</script>