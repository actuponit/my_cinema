import type { UseFetchOptions } from 'nuxt/app'

export async function useUploadImages(images: FileList) {
  const { getToken } = useApollo()
  const token = await getToken()
  const form = new FormData()
  console.log(token)
  for (let index = 0; index < images.length; index += 1) {
    let image = images.item(index)
    if (image)
      form.append('images[]', image)
  }
  return useFetch('/uploads', {
    method: 'POST',
    body: form,
    headers: {
      'Authorization': 'bearer '+token,
    },
    $fetch: useNuxtApp().$api
  })
}
