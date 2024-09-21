import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./app.vue",
    "./components/**/*.{js,vue,ts}",
    "./layouts/**/*.vue",
    "./pages/**/*.vue",
    "./plugins/**/*.{js,ts}",
    "./nuxt.config.{js,ts}",
  ],
  theme: {
    extend: {
      colors: {
        blue: {
          '50': '#eff8ff',
          '100': '#dff0ff',
          '200': '#b7e2ff',
          '300': '#77ccff',
          '400': '#2eb4ff',
          '500': '#16a6fc',
          '600': '#007ad1',
          '700': '#0060a9',
          '800': '#01528b',
          '900': '#074473',
          '950': '#052b4c',
        },
        yellow: "#f7c600",
        black: "#151717",
        primary: "#151717",
        red: "#f54747",
      },
    },
  },
  safelist: ["blue", "yellow", "black", "primary", "red"],
};


export default config;