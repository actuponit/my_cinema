<template>
  <UModal :ui="{ divide: 'divide-y divide-gray-100 dark:divide-gray-800' }">
    <UCard>
      <template #header>
        <div class="flex justify-between items-center">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
            Add actor
          </h3>
          <UButton color="gray" variant="ghost" icon="i-heroicons-x-mark-20-solid" @click="closeModal" />
        </div>
      </template>
      <p class="mb-3">Please selec an actor:</p>
      <USelectMenu 
        v-model="actors"
        name="actors"
        id="actors"
        :options="acotrsOptions"
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
          <UButton color="primary" @click="addActor">
            Add actors
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

const awhere = {is_director: {_eq: false}, _or: [{first_name: {_ilike: "%%"}}, {last_name: {_ilike: "%%"}}]}
const { refetch:refetchActors, loading:loadingActors, result:actorsResult } = useQuery(MOVIE_CASTS, {where: awhere})
const arg = ref('')
const acotrsOptions = ref([])
const actors = ref<number>(-1)

watch(() => arg, async () => {
  console.log(arg.value)
  if (awhere._or && awhere._or[0] && awhere._or[0].first_name)
    awhere._or[0].first_name._ilike = `%${arg.value}%`
  if (awhere._or && awhere._or[1] && awhere._or[1].last_name)
    awhere._or[1].last_name._ilike = `%${arg.value}%`
  acotrsOptions.value = []
  await refetchActors({where: awhere})
  if (actorsResult.value)
    acotrsOptions.value = actorsResult.value.casts.map((cast: any) => ({id: cast.id, name: `${cast.first_name} ${cast.last_name}`}))
}, {immediate: true})

const emit = defineEmits(['close'])
const closeModal = () => {
  emit('close')
}

const {executeInsert: assignMovie, onDone: assignSuccessfull} = useAssignMovie();
assignSuccessfull(() => {
  useToast().add({
    title: "Successful Operation",
    description: "Successfully added actor to crew!",
    color: 'green'
  })
  emit('close')
})

const addActor = async () => {
  if (actors.value !== -1)
    await assignMovie({cast_id: actors.value, movie_id: props.id})
}

</script>