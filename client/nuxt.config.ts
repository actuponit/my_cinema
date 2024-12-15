// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2024-04-03',
  devtools: { enabled: false },
  css: ['@/assets/css/tailwind.css'],
  modules: [
    '@nuxt/ui',
    '@nuxtjs/apollo',
  ],
  apollo: {
    clients: {
      default: {
        httpEndpoint: 'https://pleasing-viper-86.hasura.app/v1/graphql',
        connectToDevTools: true,
      }
    },
  },
  colorMode: {
    preference: 'system',
    fallback: 'dark',
    classSuffix: '',
  },
  ui: {
    safelistColors: ['blue', 'yellow', 'black', 'primary', 'red']
  },
  nitro: {
    devProxy: { '/api': `https://postman-echo.com` },
  },
})
