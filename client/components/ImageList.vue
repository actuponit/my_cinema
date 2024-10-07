<template>
  <div>
    <div v-if="images.length > 0" class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
      <div v-for="(image, index) in images" :key="index" class="relative group">
        <img :src="displayImage(image)" alt="actor image" class="w-full h-40 object-cover rounded-lg shadow-md" />
        <div v-if="id && id !== ''" class="absolute inset-0 bg-black bg-opacity-50 opacity-0 group-hover:opacity-100 transition-opacity duration-200 flex flex-col items-stretch px-3 justify-center rounded-lg">
          <UButton
            v-if="featured !== image"
            variant="solid"
            size="sm"
            icon="i-heroicons-star"
            class="mb-4"
            @click="onFeatured(image)"
          >
            Featured
          </UButton>
          <UButton
            color="red"
            variant="solid"
            icon="i-heroicons-trash"
            size="sm"
            @click="deleteImage(image)"
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
import { MOVIE_BYID } from '~/graphql/queries/movies';
import ConfirmModal from './ConfirmModal.vue';
import { MOVIE_THUMBNAIL_DELETE } from '~/graphql/mutations/movie';

const props = defineProps({
  initialImages: {
    type: Array,
    default: () => []
  },
  id: String,
  featured: String
})
const modal = useModal()
const images = ref(props.initialImages)
const {mutate, onDone: deleteThumb} = useMutation(MOVIE_THUMBNAIL_DELETE)

const emit = defineEmits(['update:featured'])

const onFeatured = (image) => {
  emit('update:featured', image)
}

deleteThumb(({data:{delete_movie_thumbnails_by_pk: {image_url}}}) => {
  useToast().add({
    title: 'Thumbnail Deleted',
    description: 'The thumbnail has been deleted successfully',
    color: 'green'
  })
  modal.close()
  images.value = images.value.filter((img) => img !== image_url)
  emit('update:featured', '')
  console.log(image_url, 'finished')
})
const deleteImage = (image) => {
  modal.open(ConfirmModal, {
    title: 'Delete Image',
    message: 'Are you sure you want to delete this image?',
    action: 'Delete',
    onAction:async () => {
      await mutate({image_url: image}, {refetchQueries: [{query: MOVIE_BYID, variables: {id: props.id}}]})
    },
    onClose: () => {
      modal.close()
    }
  })
}
</script>