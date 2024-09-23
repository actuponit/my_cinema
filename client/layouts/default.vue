<template>
    <div class="flex min-h-screen bg-black">
      <!-- Sidebar -->
      <div
        :class="[
          'transition-all duration-300 ease-in-out h-1 border-r border-gray-700',
          isOpen ? 'w-64' : 'w-20'
        ]"
      />
      <div
        :class="[
          'transition-all duration-300 ease-in-out h-screen fixed bg-black-950 border-r border-gray-700',
          isOpen ? 'w-64' : 'w-20'
        ]"
      >
        <div class="flex items-center justify-between p-4 border-b border-gray-700">
          <h1
            :class="[
              'font-bold text-gray-100 transition-all duration-300 ease-in-out',
              isOpen ? 'text-xl' : 'text-sm'
            ]"
          >
            {{ isOpen ? 'CineTix' : '' }}
          </h1>
          <button
            @click="toggleSidebar"
            class="p-2 rounded-md hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-gray-600 text-gray-400 hover:text-gray-100"
          >
            <MenuIcon v-if="!isOpen" class="w-6 h-6" />
            <XIcon v-else class="w-6 h-6" />
          </button>
        </div>
  
        <nav class="mt-6 flex flex-col justify-between h-[calc(100%-6.5rem)]">
          <div>
            <div v-for="(section, index) in mainSections" :key="index" class="mb-4">
              <button
                :class="['flex items-center w-full px-4 py-2 text-left text-gray-400 hover:bg-gray-700 hover:text-gray-100 focus:outline-none', isOpen?'justify-between':'justify-center']"
              >
                <div class="flex items-center">
                  <component :is="section.icon" class="w-5 h-5 mr-3" />
                  <span v-if="isOpen" class="text-sm font-medium">{{ section.title }}</span>
                </div>
              </button>
            </div>
          </div>
  
          <div class="mt-auto">
            <div v-for="(section, index) in userSections" :key="index" class="mb-4">
              <button
                @click="toggleSection(index, 'user')"
                :class="['flex items-center w-full px-4 py-2 text-left text-gray-400 hover:bg-gray-700 hover:text-gray-100 focus:outline-none', isOpen?'justify-between':'justify-center']"
              >
                <div class="flex items-center">
                  <component :is="section.icon" class="w-5 h-5 mr-3" />
                  <span v-if="isOpen" class="text-sm font-medium">{{ section.title }}</span>
                </div>
                <ChevronDownIcon
                  v-if="isOpen"
                  :class="[
                    'w-4 h-4 transition-transform duration-200',
                    section.isOpen ? 'transform rotate-180' : ''
                  ]"
                />
              </button>
              <transition
                enter-active-class="transition duration-100 ease-out"
                enter-from-class="transform scale-95 opacity-0"
                enter-to-class="transform scale-100 opacity-100"
                leave-active-class="transition duration-75 ease-in"
                leave-from-class="transform scale-100 opacity-100"
                leave-to-class="transform scale-95 opacity-0"
              >
                <ul v-if="isOpen && section.isOpen" class="mt-2 space-y-2">
                  <li v-for="(item, itemIndex) in section.items" :key="itemIndex">
                    <a
                      href="#"
                      class="block px-4 py-2 text-sm text-gray-400 hover:bg-gray-700 hover:text-gray-100"
                    >
                      {{ item }}
                    </a>
                  </li>
                </ul>
              </transition>
            </div>
  
            <!-- User Profile Section -->
            <div class="border-t border-gray-700 pt-4 pb-3 px-4">
              <div class="flex items-center">
                <div class="flex-shrink-0">
                  <img class="h-10 w-10 rounded-full" src="/blank-profile-picture.webp" alt="User avatar" />
                </div>
                <div v-if="isOpen" class="ml-3">
                  <div class="text-base font-medium leading-none text-gray-100">John Doe</div>
                  <div class="text-sm font-medium leading-none text-gray-400 mt-1">john@example.com</div>
                </div>
              </div>
            </div>
          </div>
        </nav>
      </div>
      <UModals />
      <UNotifications />
      <!-- Main Content -->
      <slot />
    </div>
  </template>
  
  <script setup>
  import { MenuIcon, XIcon, ChevronDownIcon, VideoIcon, HomeIcon, UserIcon, BookmarkIcon, StarIcon } from 'lucide-vue-next'
  import { ref } from 'vue'
  const isOpen = ref(true)
  
  const mainSections = ref([
    {
      title: 'Home',
      icon: HomeIcon,
    },
    {
      title: 'Movies',
      icon: VideoIcon,
    },
    {
      title: 'BookMarks',
      icon: BookmarkIcon,
    },
    {
      title: 'Ratings',
      icon: StarIcon,
    }
  ])
  
  const userSections = ref([
    {
      title: 'Account',
      icon: UserIcon,
      isOpen: false,
      items: ['Log out', 'Profile']
    },
  ])
  
  const toggleSidebar = () => {
    isOpen.value = !isOpen.value
  }
  
  const toggleSection = (index) => {
    userSections.value[index].isOpen = !userSections.value[index].isOpen
  }
  </script>