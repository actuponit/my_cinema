<template>
    <div class="container mx-auto p-6 flex-1">
    	<h1 class="text-3xl font-bold mb-6">Add New Cast Member</h1>
			<UForm :state="values" @submit="onSubmit">
				<div class="p-6">
					<div class="grid gap-6">
						<UFormGroup label="Is the preson a director" help="The default is an actor" name="isDirector">
							<UToggle v-model="isDirector" />
						</UFormGroup>
            <UFormGroup label="First Name" name="firstname" v-bind="firstnameProps" class="space-y-4">
              <UInput name="firstname" v-model="firstname" type="text" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50" />
            </UFormGroup>
            <UFormGroup label="Last Name" name="lastname" v-bind="lastnameProps" class="space-y-4">
              <UInput name="lastname" v-model="lastname" type="text" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50" />
            </UFormGroup>
            <UFormGroup label="Short Bio" name="bio" v-bind="bioProps" class="space-y-4">
              <UTextarea name="bio" :rows="5" v-model="bio" type="text" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50" />
            </UFormGroup>
            <UFormGroup label="Photo" :error="photoProps.error" help="The image must be less than 5mb and an image format" >
              <UInput type="file" accept="image/*" name="photo" class="my-3" id="photo" @change="photo=$event"/>
            </UFormGroup>
						<button type="submit" class="inline-flex col-span-1 ml-auto justify-center py-2 px-4 border justify-self-end border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
							Save Cast
						</button>
					</div>
				</div>
			</UForm>

		</div>
    
</template>
<script setup lang="ts">

import { useForm } from 'vee-validate'
import * as yup from 'yup'

const schema = yup.object().shape({
	firstname: yup.string().required("First name of the actor is required"),
	lastname: yup.string().required("Last name of the actor is required"),
	photo: yup.mixed().required("Photo of the actor is required").test('photo_size', "Photo too large", (value) => {
		if (!value) return true
		let p = (value as FileList)[0]
		if (p.size <= (5 * (1 << 20))) return true
		console.log("Photo loading problem, ", value, p)
		return false
	}),
	isDirector: yup.boolean().required(),
});

const {defineField, handleSubmit, values } = useForm({
	validationSchema: schema,
})
const nuxtUiConfig = {
	props: (state: { errors: any[]; }) => ({
		error: state.errors[0]
	}),
};
const [firstname, firstnameProps] = defineField('firstname', nuxtUiConfig)
const [lastname, lastnameProps] = defineField('lastname', nuxtUiConfig)
const [photo, photoProps] = defineField('photo', nuxtUiConfig)
const [bio, bioProps] = defineField('bio', nuxtUiConfig)
const [isDirector, isDirectorProps] = defineField('isDierector', nuxtUiConfig)

const onSubmit = handleSubmit((values) => {
	console.log(values)
	// Here you would typically send the data to your backend
})
</script>