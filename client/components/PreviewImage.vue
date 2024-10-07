<template>
  <div>
    <div v-if="previewImages.length > 0" class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
      <div v-for="(image, index) in previewImages" :key="index" class="relative">
        <img :src="(image as string)" alt="Preview" class="w-full h-40 rounded-lg shadow-md" />
      </div>
    </div>
    <p v-else class="text-gray-500 dark:text-gray-400">No images selected</p>
  </div>
</template>

<script setup lang="ts">

const props = defineProps({
  files: {
    type: Object as PropType<FileList|null>,
    default: () => []
  }
})
const previewImages = ref<Array<string | ArrayBuffer>>([])

const handleFileChange = () => {
  previewImages.value = []
  if (!props.files) return
  for (let i = 0; i < props.files.length; i++) {
    const file = props.files[i]
    const reader = new FileReader()
  
    reader.onload = (e) => {
      if (e.target && e.target.result) {
        previewImages.value.push(e.target.result)
      }
    }
  
    reader.readAsDataURL(file)
  }
}

watch(() => props.files, handleFileChange, { immediate: true })
</script>