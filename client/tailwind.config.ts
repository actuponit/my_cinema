import type { Config } from "tailwindcss";
import plugin from "tailwindcss/plugin";

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
        black: {
          '50': '#ebf8ff',
          '100': '#dbf0ff',
          '200': '#bfe3ff',
          '300': '#98cfff',
          '400': '#6fafff',
          '500': '#4e8fff',
          '600': '#2e69fd',
          '700': '#2254e0',
          '800': '#1f49b4',
          '900': '#22438d',
          '950': '#080f21',
        },
        red: {
          '50': '#fef2f2',
          '100': '#ffe1e1',
          '200': '#ffc8c8',
          '300': '#ffa2a2',
          '400': '#fc6d6d',
          '500': '#f54747',
          '600': '#e22020',
          '700': '#be1717',
          '800': '#9d1717',
          '900': '#821a1a',
          '950': '#470808',
        },
        "parimary-bg":"#080f21",
        "secondary-bg":"#1f2937",
        yellow: {
          '50': '#fffee7',
          '100': '#fffec1',
          '200': '#fff886',
          '300': '#ffed41',
          '400': '#ffdc0d',
          '500': '#f7c600',
          '600': '#d19500',
          '700': '#a66a02',
          '800': '#89530a',
          '900': '#74440f',
          '950': '#442304',
        },
      
      
      },
    },
  },
  darkMode: "class",
  safelist: ["blue", "yellow", "black", "primary", "red"],
  plugins: [
    plugin(function({ addUtilities }) {
      const newUtilities = {
        '.scrollbar-thin': {
          scrollbarWidth: 'thin',
          scrollbarColor: '#16a6fc',
          '&::-webkit-scrollbar': {
            width: '8px',
          },
          '&::-webkit-scrollbar-track': {
            backgroundColor: '#f1f1f1',
          },
          '&::-webkit-scrollbar-thumb': {
            backgroundColor: '#16a6fc',
            borderRadius: '20px',
          },
          '&::-webkit-scrollbar-thumb:hover': {
            backgroundColor: '#555',
          },
        },
      }
      addUtilities(newUtilities)
    })
  ],
};


export default config;