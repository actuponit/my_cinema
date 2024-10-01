import { CAST_MOVIE_DELETE } from "~/graphql/mutations/cast";
import { CAST_QUERY_BYID } from "~/graphql/queries/casts";

export function useCastDeleteMovie(id: string) {
  const { mutate, loading, onDone } = useMutation(CAST_MOVIE_DELETE, { refetchQueries: [{
    query: CAST_QUERY_BYID,
    variables: { id }
  }]
});

  const executeDelete = async (cast_id: string, movie_id: string) => {
    try {
      console.log('cast_id', cast_id);
      console.log('movie', movie_id, typeof movie_id);
      const response = await mutate({ cast_id, movie_id });
      console.log('response', response);
      
      // return response?.data.login.user;
    } catch (err) {
      throw createError((err as Error).message);
    }
  }
  return { loading, executeDelete, onDone };
}