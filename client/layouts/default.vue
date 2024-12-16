<template>
  <div class="flex min-h-screen bg-black">
    <!-- Sidebar -->
    <div :class="[
      'transition-all duration-300 ease-in-out h-1 border-r border-gray-700',
      isOpen ? 'w-64' : 'w-20'
    ]"></div>
    <div :class="[
      'transition-all duration-300 ease-in-out h-screen overflow-y-scroll fixed bg-black-950 border-r border-gray-700',
      isOpen ? 'w-64' : 'w-20'
    ]">
      <div class="flex items-center justify-between p-4 border-b border-gray-700">
        <h1 :class="[
          'font-bold text-gray-100 transition-all duration-300 ease-in-out',
          isOpen ? 'text-xl' : 'text-sm'
        ]">
          {{ !isOpen ? '' : role === 'cinema' ? 'Admin' : 'CineTix' }}
        </h1>
        <button @click="toggleSidebar"
          class="p-2 rounded-md hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-gray-600 text-gray-400 hover:text-gray-100">
          <MenuIcon v-if="!isOpen" class="w-6 h-6" />
          <XIcon v-else class="w-6 h-6" />
        </button>
      </div>

      <nav class="mt-6 flex flex-col justify-between h-[calc(100%-6.5rem)]">
        <div>
          <div v-for="(section, index) in mainSections" :key="section.link" class="mb-4">
              <NuxtLink :to="section.link" active-class="text-primary"
                :class="['flex items-center w-full px-4 py-2 text-left text-gray-400 hover:bg-gray-700 hover:text-gray-100 focus:outline-none', isOpen ? 'justify-between' : 'justify-center']">
                <ClientOnly>
                  <div class="flex items-center">
                      <component :key="section.link" :is="section.icon" class="w-5 h-5 mr-3" />
                    <span v-if="isOpen" class="text-sm font-medium">{{ section.title }}</span>
                  </div>
                </ClientOnly>
              </NuxtLink>
            </div>
        </div>

        <div class="mt-auto">
          <UPopover mode="hover" :popper="{ offsetDistance: 0, placement: 'top-end' }">
            <button
              :class="['flex items-center w-full px-4 py-2 text-left text-gray-400 hover:bg-gray-700 hover:text-gray-100 focus:outline-none', isOpen ? 'justify-between' : 'justify-center']">
              <div class="flex items-center">
                <UserIcon class="w-5 h-5 mr-3" />
                <span v-if="isOpen" class="text-sm font-medium">Account</span>
              </div>
            </button>
            <template #panel>
              <div class="mb-4">
                <ul class="mt-2 space-y-2">
                  <li v-for="(item, itemIndex) in items" :key="itemIndex">
                    <a @click="clearUserFunc(item.title)" :href="item.link"
                      :class="['block py-2 text-sm text-gray-400 hover:bg-gray-700 hover:text-gray-100', isOpen ? 'px-20' : 'px-8']">
                      {{ (item as SectionItem).title }}
                    </a>
                  </li>
                </ul>
              </div>
            </template>
          </UPopover>

          <!-- User Profile Section -->

          <div v-if="name && email" class="border-t border-gray-700 pt-4 pb-3 px-4">
            <div class="flex items-center">
              <div class="flex-shrink-0">
                <img class="h-10 w-10 rounded-full" src="/blank-profile-picture.webp" alt="User avatar" />
              </div>
              <div v-if="isOpen" class="ml-3">
                <div class="text-base font-medium leading-none text-gray-100">{{ name }}</div>
                <div class="text-sm font-medium leading-none text-gray-400 mt-1">{{ email }}</div>
              </div>
            </div>
          </div>
        </div>
      </nav>
    </div>
    <!-- Main Content -->
    <slot></slot>
    <UModals />
    <UNotifications />
  </div>
</template>

<script setup lang="ts">
import { MenuIcon, XIcon, VideoIcon, HomeIcon, UserIcon, BookmarkIcon, StarIcon, TicketIcon, UserPlus2Icon, VideotapeIcon, TimerIcon, } from 'lucide-vue-next'
import { ref } from 'vue'
const isOpen = ref(true)

const mainSections = ref([
  {
    title: 'Home',
    icon: HomeIcon,
    link: '/'
  },
  {
    title: 'Movies',
    icon: VideoIcon,
    link: '/movies'
  },
  {
    title: 'BookMarks',
    icon: BookmarkIcon,
    link: '/bookmarks'
  },
  {
    title: 'Reviews',
    icon: StarIcon,
    link: '/reviews'
  },
  {
    title: 'Charts',
    icon: VideoIcon,
    link: '/charts'
  },
  {
    title: 'Tickets',
    icon: TicketIcon,
    link: '/tickets'
  }
])

const name = ref<string>();
const email = ref<string>();
const role = ref<string>();

interface SectionItem {
  title: string;
  link: string;
}

const items = ref<SectionItem[]>([{
  title: 'Log in',
  link: '/login'
}, {
  title: 'Sign up',
  link: '/signup'
}
])

const admin = [{
  title: 'Casts',
  icon: UserIcon,
  link: '/admin/actors'
},
{
  title: 'Create Casts',
  icon: UserPlus2Icon,
  link: '/admin/actors/create'
},
{
  title: 'Schedules',
  icon: TimerIcon,
  link: '/admin/schedules'
},
{
  title: 'Create Movie',
  icon: VideotapeIcon,
  link: '/admin/movies/create'
}]

const toggleSidebar = () => {
  isOpen.value = !isOpen.value
}

const { user, clearUser } = useUser();

const clearUserFunc = (title: string) => {
  if (title === 'Log out') {
    clearUser();
    useToast().add({
      title: 'Logged out',
      description: 'You have been logged out',
      color: 'green'
    })
  }
}
if (user.value) {
  items.value = [{
    title: 'Log out',
    link: '/'
  }, {
    title: 'Profile',
    link: '/profile'
  }]
}

name.value = user.value?.first_name + " " + user.value?.last_name;
email.value = user.value?.email;
role.value = user.value?.role;
if (role.value === 'cinema') {
  mainSections.value.pop()
  mainSections.value = [...mainSections.value, ...admin]
}
</script>

<style>
/* Custom scrollbar styles */
::-webkit-scrollbar {
  width: 8px;
}

::-webkit-scrollbar-track {
  background: #1a1a1a;
}

::-webkit-scrollbar-thumb {
  background-color: #6a6a6a;
  border-radius: 10px;
  border: 2px solid #1a1a1a;
}

::-webkit-scrollbar-thumb:hover {
  background-color: #2eb4ff;
}
</style>