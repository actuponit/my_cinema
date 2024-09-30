import gql from 'graphql-tag'

export const CAST_INSERT = gql`
  mutation MyMutation($object: casts_insert_input!) {
    insert_casts_one(object: $object) {
      bio
      first_name
      last_name
      photo_url
      id
      is_director
    }
  }
` 
export const ASSIGN_MOVIE = gql`
  mutation MyMutation($object: crew_insert_input!) {
    insert_crew_one(object: $object) {
      cast_id
      movie_id
    }
}
`
