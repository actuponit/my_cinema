import type { UseFetchOptions } from 'nuxt/app'

export function useUploadImages(images: File[]){
  return useFetch('/uploads', {
    method: 'POST',
    body: {
      'images[]': images
    },
    headers: {
        'Content-Type': 'multipart/form-data'
    },
    $fetch: useNuxtApp().$api
  })
}
