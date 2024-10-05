import { ref, readonly } from 'vue'
import { useStorage } from '@vueuse/core'

const user = ref<User | null>(null)

export function useUser() {
  const storedUser = useStorage<string>('user', null)

  if (storedUser.value && !user.value) {
    user.value = JSON.parse(storedUser.value)
  }

  const setUser = (newUser: User) => {
    user.value = newUser
    storedUser.value = JSON.stringify(newUser)
  }

  const { onLogout } = useApollo()
  const clearUser = () => {
    onLogout();
    user.value = null
    storedUser.value = null
  }

  return {
    user: readonly(user),
    setUser,
    clearUser
  }
}