<template>
    <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-gray-900 via-primary-900 to-black-950">
      <div class="bg-white bg-opacity-10 p-8 rounded-xl shadow-2xl backdrop-blur-md w-full max-w-md">
        <h2 class="text-3xl font-bold text-center text-white mb-8">Join CineTix</h2>
        <div class="space-y-6">
          <UForm :state="values" :schema="schema" @submit="onSubmit">
            <UFormGroup name="firstname" label="First Name" v-bind="firstnameProps" class="my-4">
              <UInput v-model="firstname" color="blue"/>
            </UFormGroup>
            
            <UFormGroup name="lastname" label="Last Name" v-bind="lastnameProps" class="my-4">
              <UInput v-model="lastname" color="blue"/>
            </UFormGroup>

            <UFormGroup name="email" label="Email" v-bind="emailProps" class="my-4">
              <UInput v-model="email" color="blue" />
            </UFormGroup>
            
            <UFormGroup name="password" label="Password" v-bind="passwordProps" class="my-4">
              <UInput v-model="password" color="blue" />
            </UFormGroup>
            
            <UFormGroup name="confirmPasswod" label="Confirm Password" v-bind="confirmPasswordProps" class="my-4">
              <UInput v-model="confirmPassword" color="blue" />
            </UFormGroup>
            <button type="submit"
              class="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 my-4">
              Sign Up
            </button>
          </UForm>
        </div>
        <p class="mt-6 text-center text-sm text-gray-300">
          Already have an account?
          <NuxtLink to="/login" class="font-medium text-primary-300 hover:text-primary-200 cursor-pointer">Sign in</NuxtLink>    
        </p>
      </div>
    </div>
</template>
  
<script setup lang="ts">
  import { useForm } from 'vee-validate';
  import * as yup from 'yup';

  const schema = yup.object({
    email: yup.string().email().required().lowercase(),
    password: yup.string().required().min(6),
    firstname: yup.string().required().min(2).matches(/^[A-Za-z]+$/, 'First name must not contain numbers'),
    lastname: yup.string().required().min(2).matches(/^[A-Za-z]+$/, 'Last name must not contain numbers'),
    confirmPassword: yup.string().oneOf([yup.ref('password')], 'Passwords must match'),
  });

  const {defineField, handleSubmit, values} = useForm({
    validationSchema: schema,
  });
  
  const onSubmit = handleSubmit((values) => {
    console.log(values);
  });

  const nuxtUiConfig = (state: { errors: any[]; }) => {
    return {
      props: {
        error: state.errors[0],
      },
    };
  };
  const [email, emailProps] = defineField('email', nuxtUiConfig);
  const [password, passwordProps] = defineField('password', nuxtUiConfig);
  const [firstname, firstnameProps] = defineField('firstname', nuxtUiConfig);
  const [lastname, lastnameProps] = defineField('lastname', nuxtUiConfig);
  const [confirmPassword, confirmPasswordProps] = defineField('confirmPassword', nuxtUiConfig);

  definePageMeta({
    layout: false,
  });
</script>