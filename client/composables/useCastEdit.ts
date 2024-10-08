import { CAST_UPDATE } from "~/graphql/mutations/cast";
import { CAST_QUERY_BYID } from "~/graphql/queries/casts";

export function useCastEdit() {
  const { mutate, loading, onDone } = useMutation(CAST_UPDATE);

  const executeUpdate = async (id: string, values: any) => {
    try {
      const response = await mutate({ id, _set: values }, {
        refetchQueries: [
          { query: CAST_QUERY_BYID, variables: { id } }
        ]
      });
      console.log('response', response);
      
      // return response?.data.login.user;
    } catch (err) {
      throw createError((err as Error).message);
    }
  }
  return { loading, executeUpdate, onDone };
}