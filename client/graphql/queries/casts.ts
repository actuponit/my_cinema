import gql from 'graphql-tag'

export const CAST_QUERY = gql`
  query MyQuery {
    casts {
      first_name
      id
      is_director
      last_name
      movies_aggregate(distinct_on: id) {
        aggregate {
          count(distinct: true)
        }
      }
      crews_aggregate(distinct_on: movie_id) {
        aggregate {
          count(distinct: true)
        }
      }
    }
  }
` 

export const CAST_QUERY_BYID = gql`
  query MyQuery($id: Int!) {
  casts_by_pk(id: $id) {
    bio
    crews(distinct_on: movie_id) {
      movie {
        published_at
        title
      }
    }
    movies(distinct_on: id) {
      published_at
      title
    }
    first_name
    id
    is_director
    last_name
    photo_url
  }
}
`
export const CAST_QUERY_MOVIE = gql`

  query MyQuery($where: movies_bool_exp!) {
  movies(where: $where) {
    title
    published_at
    id
  }
}
`