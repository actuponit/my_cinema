export default defineNuxtRouteMiddleware(async (to, from) => {
	if (to.fullPath === '/logout') {
    const { onLogout } = useApollo();
    await onLogout();
    return navigateTo('/');
  }
  if (to.fullPath.startsWith('/admin')) {
    const { getToken } = useApollo();
    const token = await getToken();
    if (!token) {
      return navigateTo('/login');
    }
    
    const { user } = useUser();
    const role = user.value?.role
    if (role && role === 'cinema') return
    if (import.meta.client)
      return navigateTo('/');
  }
})