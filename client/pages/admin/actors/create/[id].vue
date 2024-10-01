<template>
  <div class="container mx-auto p-6 flex-1">
    <h1 class="text-3xl font-bold mb-6">Edit Cast Member</h1>
    <div class="p-6">
      <div v-if="error || !data?.casts_by_pk">
        <ErrorComponent title="Couldn't find the cast" :message="error?.message || ''" />
      </div>
      <form v-else method="post" @submit="onSubmit">
        <div class="grid gap-6">
          <UFormGroup label="Is the preson a director" help="The default is an actor" name="isDirector">
            <UToggle v-model="isDirector" size="2xl" class="mt-3" />
          </UFormGroup>
          <UFormGroup label="First Name" name="firstname" v-bind="firstnameProps" class="space-y-4">
            <UInput name="firstname" v-model="firstname" type="text" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50" />
          </UFormGroup>
          <UFormGroup label="Last Name" name="lastname" v-bind="lastnameProps" class="space-y-4">
            <UInput name="lastname" v-model="lastname" type="text" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50" />
          </UFormGroup>
          <UFormGroup label="Short Bio" name="bio" v-bind="bioProps" class="space-y-4">
            <UTextarea name="bio" id="bio" v-model="bio" :rows="5" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50" />
          </UFormGroup>
          <UFormGroup label="Photo" :error="photoProps.error" help="The image must be less than 5mb and an image format, And If you choose a new image the old one will be deleted" >
            <UInput type="file" accept="image/*" name="photo" class="my-3" id="photo" @change="photo=$event"/>
          </UFormGroup>
          <div class="flex justify-between gap-4">
            <div>
              <p class="text-sm text-gray-500 mb-2">Preview of the selected image</p>
              <PreviewImage :files="photo" />
            </div>
            <div>
              <p class="text-sm text-gray-500 mb-2">Previous image</p>
              <ImageList v-if="prevPhoto" :initial-images="[prevPhoto]" />
            </div>
          </div>

          <UButton :loading="loading" @click="onSubmit" type="submit" class="inline-flex col-span-1 ml-auto justify-center py-2 px-4 border justify-self-end border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
            Update Cast
          </UButton>
        </div>
      </form>
    </div>

  </div>
  
</template>
<script setup lang="ts">

import { useForm } from 'vee-validate'
import * as yup from 'yup'
import { CAST_QUERY_BYID } from '~/graphql/queries/casts';

const {id} = useRoute().params;
const router = useRouter();

const schema = yup.object().shape({
firstname: yup.string().required("First name of the actor is required"),
lastname: yup.string().required("Last name of the actor is required"),
bio: yup.string().required("Bio is required"),
photo: yup.mixed().nullable().test('photo_size', "Photo too large", (value) => {
  if (value === null || (value as FileList).length === 0) return true
  let p = (value as FileList)[0]
  if (p.size <= (5 * (1 << 20))) return true
  console.log("Photo loading problem, ", value, p)
  return false
}),
isDirector: yup.boolean().required(),
});

const { result:data, error } = useQuery<{ casts_by_pk: Cast }>(CAST_QUERY_BYID, {id})
// console.log("bio", data.value?.casts_by_pk.bio)
console.log(data.value)
const prevPhoto = ref<string | undefined>(data.value?.casts_by_pk.photo_url)
const initialValues = ref(
  data.value?.casts_by_pk ?({
    firstname: data.value?.casts_by_pk.first_name, 
    lastname: data.value?.casts_by_pk.last_name,
    bio: (data.value?.casts_by_pk.bio),
    isDirector: data.value?.casts_by_pk.is_director,
    photo: null
  }) : {
    firstname: '', 
    lastname: '',
    bio:'',
    isDirector:  false,
    photo: null
  })

const {defineField, handleSubmit, } = useForm({
validationSchema: schema,
initialValues: initialValues.value
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
const [isDirector, isDirectorProps] = defineField('isDirector', nuxtUiConfig)

const { executeUpdate, loading, onDone } = useCastEdit();

const toast = useToast();
onDone(() => {
  toast.add({
    color: 'green',
    title: 'Actor Updated',
    description: 'Your actor has been updated'
  })
  router.replace('/admin/actors/'+id)
})
const onSubmit = handleSubmit(async (values) => {
try {
  if (values.photo) {
    const { data, status } = await useUploadImages(values.photo);
    console.log(data.value)
    if (status.value === 'success') {
      console.log("Photo uploaded successfully")
    }
    if (!data.value) return
    const castedData = Array.isArray(data.value) ? data.value.map(item => ({ image_url: item.image_url })) : [];
    await executeUpdate(id as string, {
      first_name: values.firstname,
      last_name: values.lastname,
      bio: values.bio,
      photo_url: castedData[0].image_url,
      is_director: values.isDirector
    })
  } else {
    await executeUpdate(id as string, {
      first_name: values.firstname,
      last_name: values.lastname,
      bio: values.bio,
      is_director: values.isDirector
    })
  }
} catch (error) {
  toast.add({
    title: "Error while updating cast",
    color: "red"
  })
  console.log("Error casting data", error)
}
})

</script>