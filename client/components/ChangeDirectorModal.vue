<template>
  <UModal :ui="{ divide: 'divide-y divide-gray-100 dark:divide-gray-800' }">
    <UCard>
      <template #header>
        <div class="flex justify-between items-center">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
            Change a director
          </h3>
          <UButton color="gray" variant="ghost" icon="i-heroicons-x-mark-20-solid" @click="closeModal" />
        </div>
      </template>
      <p class="mb-3">Please selec a director:</p>
      <USelectMenu 
        v-model="director"
        :options="directorOptions"
        value-attribute="id" 
        option-attribute="name"
        v-model:query="arg"
        searchable
        :loading="loadingActors"
        :searchableLazy="true"
        :debounce="400"
      />
      <template #footer>
        <div class="flex justify-end space-x-3">
          <UButton color="gray" variant="outline" @click="closeModal">
            Cancel
          </UButton>
          <UButton color="primary" @click="updateDirector">
            Change director
          </UButton>
        </div> 
      </template>
    </UCard>
  </UModal>
</template>
<script setup lang="ts">
import { MOVIE_CASTS } from '~/graphql/queries/movies';

const props = defineProps({
  id: String,
})

const where = {is_director: {_eq: true}, _or: [{first_name: {_ilike: "%%"}}, {last_name: {_ilike: "%%"}}]}
const { refetch:refetchActors, loading:loadingActors, result:directorResult } = useQuery(MOVIE_CASTS, {where: where})
const arg = ref('')
const directorOptions = ref([])
const director = ref<number>(-1)

watch(() => arg, async () => {
  if (where._or && where._or[0] && where._or[0].first_name)
    where._or[0].first_name._ilike = `%${arg.value}%`
  if (where._or && where._or[1] && where._or[1].last_name)
    where._or[1].last_name._ilike = `%${arg.value}%`
  directorOptions.value = []
  await refetchActors({where: where})
  if (directorResult.value)
    directorOptions.value = directorResult.value.casts.map((cast: any) => ({id: cast.id, name: `${cast.first_name} ${cast.last_name}`}))
}, {immediate: true})

const emit = defineEmits(['close'])
const closeModal = () => {
  emit('close')
}

const {executeUpdate, onDone} = useUpdateMovie();
onDone(() => {
  useToast().add({
    title: "Successful Operation",
    description: "Successfully changed director of this movie!",
    color: 'green'
  })
  emit('close')
})

const updateDirector = async () => {
  if (director.value !== -1)
    await executeUpdate(props.id as string, {director: director.value})
}

</script>