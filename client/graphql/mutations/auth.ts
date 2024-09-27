import gql from 'graphql-tag'

export const LOGIN_MUTATION = gql`
    mutation login($email: String!, $password: String!) {
      login(email: $email, password: $password) {
        accessToken
        user {
          id
          email
          first_name
          last_name
          role
        }
      }
    }
` 

export const SIGNUP_MUTATION = gql`
  mutation SignUp($user: UserInput!) {
    signup(arg1: $user){
      accessToken
      user  {
        id
        email
        first_name
        last_name
        role
      }
    }
  }
` 