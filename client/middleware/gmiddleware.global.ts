export default defineNuxtRouteMiddleware(async (to, from) => {
	if (to.fullPath === '/logout') {
    const { onLogout } = useApollo();
    onLogout();
    return navigateTo('/');
  }
  if (to.fullPath.startsWith('/admin')) {
    const { getToken } = useApollo();
    const token = await getToken();
    console.log('token', token);
    if (!token) {
      return navigateTo('/login');
    }
    if (import.meta.client) {
      const { user } = useUser();
      const role = user.value?.role
      console.log('role', role);
      if (role && role === 'cinema') return
      return navigateTo('/');
    }
  }
})