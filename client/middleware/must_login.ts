export default defineNuxtRouteMiddleware(async (to, from) => {
	const { getToken } = useApollo();
	const token = await getToken();
	if (!token) {
    useToast().add({
      title: 'Unauthenticated',
      description: 'You need to login to access this page',
      color: 'red'
    })
		return navigateTo('/login');
	}
})