import { useMutation } from '@vue/apollo-composable';
import { SIGNUP_MUTATION } from '~/graphql/mutations/auth';

export default function useSignUp() {
  const { mutate, loading } = useMutation(SIGNUP_MUTATION);

	const executeSignUp = async (user: User) => {
		const { onLogin,} = useApollo();
    try {
			const response = await mutate({ user });
      if (!response?.data?.signup) {
        throw createError('Invalid response');
      }
			onLogin(response?.data.signup.accessToken);
			return response?.data.signup.user;
    } catch (err) {
      throw createError((err as Error).message);
    }
  }
  return { executeSignUp, loading, };
}