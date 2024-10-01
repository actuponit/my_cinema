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

export const CAST_UPDATE = gql`
  mutation MyMutation($id: Int!, $_set: casts_set_input!) {
    update_casts_by_pk(pk_columns: {id: $id}, _set: $_set) {
      bio
      first_name
      id
      is_director
      last_name
    }
  }
`

export const CAST_DELETE = gql`
  mutation MyMutation($id: Int!) {
    delete_casts_by_pk(id: $id) {
      id
    }
  }
`
