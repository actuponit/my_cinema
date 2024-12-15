export default defineNuxtPlugin(async (nuxtApp) => {
  const { getToken } = useApollo();
  const token = await getToken();

  const api = $fetch.create({
    baseURL: 'https://my-cinema-1.onrender.com',
    onRequest({ request, options, error }) {
      if (token) {
        const headers = options.headers ||= new Headers()
        if (Array.isArray(headers)) {
          headers.push(['Authorization', `Bearer ${token}`])
        } else if (headers instanceof Headers) {
          headers.set('Authorization', `Bearer ${token}`)
        } else {
          (headers as Record<string, string>).Authorization = `Bearer ${token}`
        }
      }
    },
    async onResponseError({ response }) {
      console.log('response', response)
    }
  })

  // Expose to useNuxtApp().$api
  return {
    provide: {
      api
    }
  }
})
