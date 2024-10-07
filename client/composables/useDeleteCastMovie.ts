import { CAST_MOVIE_DELETE } from "~/graphql/mutations/cast";
import { CAST_QUERY_BYID } from "~/graphql/queries/casts";
import { MOVIE_BYID } from "~/graphql/queries/movies";

export function useCastDeleteMovie() {
  const { mutate, loading, onDone } = useMutation(CAST_MOVIE_DELETE);

  const executeDelete = async (cast_id: string, movie_id: string) => {
    try {
      const response = await mutate({ cast_id, movie_id }, { refetchQueries: [{
        query: CAST_QUERY_BYID,
        variables: { id:cast_id }
      }, {
        query: MOVIE_BYID,
        variables: { id:movie_id }
      }]
    });
      
      // return response?.data.login.user;
    } catch (err) {
      throw createError((err as Error).message);
    }
  }
  return { loading, executeDelete, onDone };
}