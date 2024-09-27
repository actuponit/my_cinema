export default defineNuxtRouteMiddleware((to, from) => {
	if (to.fullPath === '/logout') {
    const { onLogout } = useApollo();
    onLogout();
    return navigateTo('/');
  }
})