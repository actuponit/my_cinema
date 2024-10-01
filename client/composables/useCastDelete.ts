import { CAST_DELETE } from "~/graphql/mutations/cast";
import { CAST_QUERY } from "~/graphql/queries/casts";

export function useCastDelete() {
  const { mutate, loading, onDone } = useMutation(CAST_DELETE, {update: (cache, { data: {delete_casts_by_pk} }) => {
    const data: { casts: { id: string }[] } | null = cache.readQuery({ query: CAST_QUERY });
    console.log('data', data);
    console.log('data', delete_casts_by_pk);
    
    if (!data) return;
    
    cache.writeQuery({
      query: CAST_QUERY,
      data: {
        casts: data.casts.filter((cast: any) => cast.id !== delete_casts_by_pk.id)
      }
    });
  }});

  const executeUpdate = async (id: string,) => {
    try {
      const response = await mutate({ id });
      console.log('response', response);
      
      // return response?.data.login.user;
    } catch (err) {
      throw createError((err as Error).message);
    }
  }
  return { loading, executeUpdate, onDone };
}