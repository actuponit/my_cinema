// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2024-04-03',
  devtools: { enabled: false },
  css: ['@/assets/css/tailwind.css'],
  modules: [
    '@nuxt/ui',
  ],
  colorMode: {
    preference: 'system',
    fallback: 'dark',
    classSuffix: '',
  },
  ui: {
    safelistColors: ['blue', 'yellow', 'black', 'primary', 'red']
  }
})