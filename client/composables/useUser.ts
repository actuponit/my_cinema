import { ref, readonly } from 'vue'
import { useStorage } from '@vueuse/core'

const user = ref<User | null>(null)

export function useUser() {
  const storedUser = useStorage('user', null, localStorage, {
    serializer: {
      read: (v) => JSON.parse(v),
      write: (v) => JSON.stringify(v),
    },
  })

  if (storedUser.value && !user.value) {
    user.value = storedUser.value
  }

  const setUser = (newUser: User) => {
    user.value = newUser
    storedUser.value = newUser
  }

  const { onLogout } = useApollo()
  const clearUser = () => {
    onLogout();
    user.value = null
    console.log('clearing user', storedUser.value)
    storedUser.value = null
    console.log('clearing use after', storedUser.value)
  }

  return {
    user: readonly(user),
    setUser,
    clearUser
  }
}