import { ref, readonly } from 'vue'
import { useCookie } from 'nuxt/app'

const user = ref<User | null>(null)

export function useUser() {
  const userCookie = useCookie('user', {
    sameSite: 'strict'
  })

  if (userCookie.value && !user.value) {
    try {
      user.value = JSON.parse(userCookie.value)
    } catch (e) {
      console.error('Failed to parse user cookie:', e)
      userCookie.value = null
    }
  }

  const setUser = (newUser: User) => {
    user.value = newUser
    userCookie.value = JSON.stringify(newUser)
  }

  const clearUser = () => {
    user.value = null
    userCookie.value = null
  }

  return {
    user: readonly(user),
    setUser,
    clearUser
  }
}