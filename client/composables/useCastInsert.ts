import { CAST_INSERT } from "~/graphql/mutations/cast";

export function useCastInsert() {
  const { mutate, loading, } = useMutation(CAST_INSERT);

  const executeInsert = async (values: any) => {
    try {
      const response = await mutate({ object: values });
      console.log('response', response);
      
      // return response?.data.login.user;
    } catch (err) {
      throw createError((err as Error).message);
    }
  }
  return { loading, executeInsert };
}