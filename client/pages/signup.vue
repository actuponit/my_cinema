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
        <h2 class="text-3xl font-bold text-center text-white mb-8">Join CineTix</h2>
        <div class="space-y-6">
          <UForm :state="values" @submit="onSubmit">
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
              <UInput v-model="password" color="blue" type="password" />
            </UFormGroup>
            
            <UFormGroup name="confirmPasswod" label="Confirm Password" v-bind="confirmPasswordProps" class="my-4">
              <UInput v-model="confirmPassword" color="blue" type="password" />
            </UFormGroup>
            <UButton type="submit" :loading="loading" size="lg" class="my-12 flex justify-center w-full text-center ">
              Sign Up
            </UButton>
          </UForm>
        </div>
        <p class="mt-6 text-center text-sm text-gray-300">
          Already have an account?
          <NuxtLink to="/login" class="font-medium text-primary-300 hover:text-primary-200 cursor-pointer">Sign in</NuxtLink>    
        </p>
      </div>
      <UNotifications />
    </div>
</template>
  
<script setup lang="ts">
  import { useForm } from 'vee-validate';
  import * as yup from 'yup';
  import useSignUp from '~/composables/useRegisterUser';

  const toast = useToast();
  const router = useRouter();
  const schema = yup.object({
    email: yup.string().email().required().lowercase(),
    password: yup.string().required().min(6),
    firstname: yup.string().required().min(2).matches(/^[A-Za-z]+$/, 'First name must not contain numbers'),
    lastname: yup.string().required().min(2).matches(/^[A-Za-z]+$/, 'Last name must not contain numbers'),
    confirmPassword: yup.string().required().oneOf([yup.ref('password')], 'Passwords must match'),
  });

  const {defineField, handleSubmit, values} = useForm({
    validationSchema: schema,
  });
  
  const { executeSignUp, loading, } = useSignUp();
  
  const onSubmit = handleSubmit(async (values) => {
    const user: User = {
      email: values.email,
      password: values.password,
      first_name: values.firstname,
      last_name: values.lastname,
    };
    try {
      const { setUser } = useUser();
      const res = await executeSignUp(user);
      setUser(res);
      toast.add({ id: 'auth', title: 'Successfull Signed Up', description: 'Account created successfully', color: 'green'});
      router.replace({path: '/'});
    } catch (err) {
      const errorMessage = (err as Error).message;
      console.log("Sign Up: ", errorMessage);
    }
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
    middleware: ['auth'],
  });
</script>