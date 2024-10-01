<template>
    <div class="w-full flex-1 min-h-full p-8">
        <div class="flex justify-between items-center">
            <h1 class="text-3xl font-bold mb-5">Casts And Crews</h1>
        </div>
        <UTable 
            class="w-full" 
            :rows="rows" 
            :loading="status"
            :empty-state="{ icon: 'i-heroicons-circle-stack-20-solid', label: error?.message || 'No items.' }"
            :loading-state="{ icon: 'i-heroicons-arrow-path-20-solid', label: 'Loading casts...' }"
            :progress="{ color: 'primary', animation: 'carousel' }" 
            :columns="columns"
            @select="onSelect"
        >
            <template #actions-data="{ row }">
                <div class="flex justify-center m-auto">
                    <button @click="onDelete(row.id, $event)" class="px-3 py-1 bg-red-600 text-white rounded-md hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2 focus:ring-offset-gray-800">
                        Delete
                    </button>
                </div>
            </template>
        </UTable>
    </div>
</template>
<script setup lang="ts">
import { CAST_QUERY } from '~/graphql/queries/casts';

const { result:data, loading:status, error } = useQuery<{ casts: Cast[] }>(CAST_QUERY,)
const router = useRouter();

const onSelect = ({ id }: {id: number}) => {
    router.push(`/admin/actors/${id}`,)
}
const rows = computed(() => {
    if (data.value) {
        return data.value?.casts.map((castInfo: Cast) => ({
            first_name: castInfo.first_name,
            last_name: castInfo.last_name,
            role: castInfo.is_director ? "Director": "Actor",
            movies: ((castInfo.movies_aggregate?.aggregate.count || 0) + (castInfo.crews_aggregate?.aggregate.count || 0)),
            id: castInfo.id
        }))
    }
})

const { executeUpdate, loading, onDone } = useCastDelete()
const toast = useToast();
const onDelete = async (id: any, event: any) => {
    event.stopPropagation(); 
    try {
        await executeUpdate(id)
    } catch (error) {
        toast.add({
            title: "Error",
            description: (error as string),
            color: "red"
        })
    }
}

onDone(() => {
    toast.add({
        title: "Successfull operation",
        description: "Cast member deleted successfuly!!",
        color: "green"
    })
})

const columns = [
    {key: "first_name", label: "First Name"},
    {key: "last_name", label: "Last Name"},
    {key: "role", label: "Role"},
    {key: "movies", label: "Number of Movies"},
    {key: "actions", label: ""},
]
</script>