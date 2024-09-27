<template>
    <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-gray-900 via-primary-900 to-black-950">
      <div class="fixed right-2 top-2">
        <UButton
          color="gray"
          variant="ghost"
          size="lg"
          to="/"
        >
          Back To Home
        </UButton>
      </div>
      <div class="bg-white bg-opacity-10 p-8 rounded-xl shadow-2xl backdrop-blur-md w-full max-w-md">
        <h2 class="text-3xl font-bold text-center text-white mb-8">Sign In to CineTix</h2>
        <UForm :state="values" @submit="onSubmit">
          <UFormGroup label="Email" name="email" v-bind="emailProps" class="my-4">
            <UInput v-model="email" :color="'blue'" trailing-icon="emailProps.error ? 'i-heroicons-exclamation-triangle-20-solid' : undefined"/>
          </UFormGroup>
          <UFormGroup label="Password" name="password" v-bind="passwordProps" class="mt-4 mb-8">
            <UInput v-model="password" :color="'blue'" type="password" />
          </UFormGroup>
          <div class="flex items-center justify-between">
            <!-- <div class="flex items-center">
              <input id="remember-me" name="remember-me" type="checkbox" class="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-300 rounded">
              <label for="remember-me" class="ml-2 block text-sm text-gray-200">Remember me</label>
            </div>
            <div class="text-sm">
              <a href="#" class="font-medium text-primary-300 hover:text-primary-200">Forgot your password?</a>
            </div> -->
          </div>
          <div>
            <UButton type="submit" size="lg" class="w-full flex justify-center mt-6" :loading="loading">
              Sign In
            </UButton>
          </div>
        </UForm>
        <p class="mt-6 text-center text-sm text-gray-300">
          Not a member?
          <NuxtLink to="/signup" class="font-medium text-primary-300 hover:text-primary-200">Sign up now</NuxtLink>
        </p>
      </div>
    </div>
    <UNotifications />
  </template>
  
<script setup lang="ts">
  import { useForm } from 'vee-validate';
  import * as yup from 'yup';

  const schema = yup.object({
    email: yup.string().email().required().lowercase().label('Email address'),
    password: yup.string().required().min(6).label('Password'),
  });

  const { defineField, handleSubmit, values } = useForm({
    validationSchema: schema,
  });

  const { loading, executeLogin } = useLogin();
  const toast = useToast();

  const router = useRouter();
  const onSubmit = handleSubmit(async (values) => {
    // console.log(values);
    try {
      const user = await executeLogin(values.email, values.password);
      if (user) {
        localStorage.setItem('name', user.first_name + " " + user.last_name);
        localStorage.setItem('email', user.email);
        localStorage.setItem('role', user.role);
      }
      toast.add({
        title: 'Login successful',
        description: 'You have successfully logged in.',
        color: 'green',
      });
      router.replace('/');
    } catch (error) {
      console.log(error);
    }
    
  });

  const nuxtUiConfig = {
    props: (state: { errors: any[]; }) => ({
      error: state.errors[0],
    }),
  };

  const [email, emailProps] = defineField('email', nuxtUiConfig);
  const [password, passwordProps] = defineField('password', nuxtUiConfig);

  definePageMeta({
    layout: false,
    middleware: ['auth'],
  });
</script>