import type { UseFetchOptions } from 'nuxt/app'

export function useUploadImages(images: File[]){
  return useFetch('/uploads', {
    method: 'POST',
    body: {
      'images[]': [] // You can replace the empty array with the actual images array
    },
    headers: {
        'Content-Type': 'multipart/form-data'
    },
    $fetch: useNuxtApp().$api
  })
}
