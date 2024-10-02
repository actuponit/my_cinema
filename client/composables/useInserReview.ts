import { CAST_INSERT } from "~/graphql/mutations/cast";
import { MOVIE_REVIEW } from "~/graphql/mutations/movie";
import { CAST_QUERY } from "~/graphql/queries/casts";
import { MOVIE_BYID } from "~/graphql/queries/movies";

export function useInsertReview(id: string) {
  const { mutate, loading, onDone } = useMutation(MOVIE_REVIEW, {
    refetchQueries: [{query: MOVIE_BYID, variables: {id}}],
  });

  const executeInsert = async (values: any) => {
    try {
      console.log("here", values)
      const response = await mutate({ object: values });  
      // return response?.data.login.user;
    } catch (err) {
      throw createError((err as Error).message);
    }
  }
  return { loading, executeInsert, onDone };
}