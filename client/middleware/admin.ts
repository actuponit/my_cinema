export default defineNuxtRouteMiddleware(async (to, from) => {
	const { user } = useUser();
  console.log('role', user);
  // if (role && role === 'cinema') return
  return navigateTo('/');
})