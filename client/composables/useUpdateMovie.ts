import { MOVIE_UPDATE } from "~/graphql/mutations/movie";
import { MOVIE_BYID } from "~/graphql/queries/movies";

export function useUpdateMovie() {
  const { mutate, loading, onDone } = useMutation(MOVIE_UPDATE);

  const executeUpdate = async (id: string, values: any) => {
    try {
      const response = await mutate({ id, _set: values }, {
        refetchQueries: [
          { query: MOVIE_BYID, variables: { id } }
        ]
      });
      // return response?.data.login.user;
    } catch (err) {
      throw createError((err as Error).message);
    }
  }
  return { loading, executeUpdate, onDone };
}