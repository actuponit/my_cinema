import { ADD_SCHEDULE } from "~/graphql/mutations/schedule";
import { MOVIE_BYID } from "~/graphql/queries/movies";

export function useAddSchedule() {
  const { mutate, loading, onDone } = useMutation(ADD_SCHEDULE)

  const executeInsert = async (values: any) => {
    try {
      const response = await mutate({ object: values }, {
        refetchQueries: [
          {query: MOVIE_BYID, variables: { id: values[0].movie.toString() }},
        ]
      });
      console.log('response', response);
      
    } catch (err) {
      throw createError((err as Error).message);
    }
  }
  return { loading, executeInsert, onDone };
}