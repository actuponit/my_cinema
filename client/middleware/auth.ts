export default defineNuxtRouteMiddleware(async (to, from) => {
	const { getToken } = useApollo();
	const token = await getToken();
	console.log('token', token);
	if (token) {
		return navigateTo('/');
	}
})