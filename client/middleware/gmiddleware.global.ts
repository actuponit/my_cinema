export default defineNuxtRouteMiddleware(async (to, from) => {
	if (to.fullPath === '/logout') {
    const { onLogout } = useApollo();
    const { clearUser } = useUser();
    onLogout();
    clearUser();
    
    return navigateTo('/');
  }
})