import { ASSIGN_MOVIE } from "~/graphql/mutations/cast";
import { CAST_QUERY, CAST_QUERY_BYID } from "~/graphql/queries/casts";

export function useAssignMovie() {
  const { mutate, loading, onDone } = useMutation(ASSIGN_MOVIE)

  const executeInsert = async (values: any) => {
    try {
      const response = await mutate({ object: values }, {
        refetchQueries: [{query: CAST_QUERY_BYID, variables: { id: values.cast_id }}]
      });
      console.log('response', response);
      
    } catch (err) {
      throw createError((err as Error).message);
    }
  }
  return { loading, executeInsert, onDone };
}