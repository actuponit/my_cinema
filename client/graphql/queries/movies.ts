export const MOVIE_CASTS = gql`
  query MyQuery($where: casts_bool_exp!) {
    casts(where: $where) {
      first_name
      last_name
      id
    }
  }
`