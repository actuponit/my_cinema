import { useMutation } from '@vue/apollo-composable';
import { LOGIN_MUTATION } from '~/graphql/mutations/auth';

export default function useSignUp() {
  const { mutate, onDone, loading, error } = useMutation(LOGIN_MUTATION);

	const executeLogin = async (email:string, password:string) => {
	const { onLogin, onLogout } = useApollo();
    try {
      const response = await mutate({ email, password });
      if (!response?.data?.login) {
        throw createError('Invalid response');
      }
      onLogin(response?.data.login.accessToken);
      return response?.data.login.user;
    } catch (err) {
      throw createError((err as Error).message);
    }
  }
  return { executeLogin, onDone, loading, error };
}