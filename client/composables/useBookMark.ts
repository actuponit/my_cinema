import { BOOKMARK_MOVIE } from "~/graphql/mutations/movie";
import { BOOKMARKS_QUERY, MOVIE_BYID } from "~/graphql/queries/movies";

export function useBookMark() {
  const { mutate, loading, onDone } = useMutation(BOOKMARK_MOVIE);

  const executeInsert = async (values: any) => {
    try {
      console.log("values", values);
      const response = await mutate(
        { object: values },
        {
          refetchQueries: [
            {
              query: MOVIE_BYID,
              variables: { id: values.movie_id, user_id: values.user },
            },
            { query: BOOKMARKS_QUERY, variables: { movie: { title: "asc" } } },
          ],
        }
      );
      console.log("response", response);

      // return response?.data.login.user;
    } catch (err) {
      throw createError((err as Error).message);
    }
  };
  return { loading, executeInsert, onDone };
}
