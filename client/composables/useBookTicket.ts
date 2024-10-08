import { BOOKMARK_MOVIE } from "~/graphql/mutations/movie";
import { BOOK_TICKET } from "~/graphql/mutations/schedule";
import { MOVIE_BYID } from "~/graphql/queries/movies";

export function useBookTicket() {
  const { mutate, loading, onDone } = useMutation(BOOK_TICKET);

  const executeInsert = async (values: any) => {
    try {
      console.log('values', values);
      const response = await mutate({ object: values });
      console.log('response', response);
      
      // return response?.data.login.user;
    } catch (err) {
      throw createError((err as Error).message);
    }
  }
  return { loading, executeInsert, onDone };
}