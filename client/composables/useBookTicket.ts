import { BOOK_TICKET } from "~/graphql/mutations/schedule";
import { TICKET_QUERY } from "~/graphql/queries/tickets";

export function useBookTicket() {
  const { mutate, loading, onDone } = useMutation(BOOK_TICKET);

  const executeInsert = async (values: any) => {
    try {
      console.log('values', values);
      const response = await mutate({ object: values }, { refetchQueries: [{ query: TICKET_QUERY  }]});
      console.log('response', response);
    } catch (err) {
      throw createError((err as Error).message);
    }
  }
  return { loading, executeInsert, onDone };
}