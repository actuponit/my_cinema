export const HOME_QUERY = gql`
query MyQuery($lte: timestamptz!, $gte: timestamptz!) {
  movies(where: {schedules: {start_time: {_lte: $lte, _gte: $gte}}}) {
    description
    featured_image
    genre
    id
    published_at
    title
    total_rating
    average_rating
    schedules(where: {start_time: {_lte: $lte, _gte: $gte}}) {
      start_time
      id
      hall
      format
      price
    }
  }
}
` 