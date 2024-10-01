import { CAST_INSERT } from "~/graphql/mutations/cast";
import { CAST_QUERY } from "~/graphql/queries/casts";

export function useCastInsert() {
  const { mutate, loading, onDone } = useMutation(CAST_INSERT, {
    refetchQueries: [{query: CAST_QUERY}]
  });

  const executeInsert = async (values: any) => {
    try {
      const response = await mutate({ object: values });
      console.log('response', response);
      
      // return response?.data.login.user;
    } catch (err) {
      throw createError((err as Error).message);
    }
  }
  return { loading, executeInsert, onDone };
}