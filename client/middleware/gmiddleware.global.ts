export default defineNuxtRouteMiddleware(async (to, from) => {
	if (to.fullPath === '/logout') {
    const { onLogout } = useApollo();
    onLogout();
    localStorage.removeItem('name');
    localStorage.removeItem('email');
    localStorage.removeItem('role');
    return navigateTo('/');
  }
})