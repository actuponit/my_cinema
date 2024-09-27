export default defineNuxtPlugin((nuxtApp) => {
	nuxtApp.hook('apollo:error', (error) => {
		console.error("From plugin", error.graphQLErrors)
		const toast = useToast()
		const errorMessages = error.graphQLErrors?.map((err) => err.message) || []
		errorMessages.forEach((message) => {
			toast.add({
				color: 'red',
				title: 'GraphQL Error',
				description: message,
				timeout: 7000,
			})
		})

		return errorMessages
	})
})