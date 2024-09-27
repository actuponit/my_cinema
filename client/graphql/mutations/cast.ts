import gql from 'graphql-tag'

export const CAST_INSERT = gql`
  mutation MyMutation($object: casts_insert_input!) {
    insert_casts_one(object: $object) {
      bio
      first_name
      last_name
      photo_url
      bio
      id
    }
  }
` 