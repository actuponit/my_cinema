export const MOVIE_CASTS = gql`
  query MyQuery($where: casts_bool_exp!) {
    casts(where: $where) {
      first_name
      last_name
      id
    }
  }
`

export const MOVIE_BYID = gql`
query MyQuery($id: Int!) {
  movies_by_pk(id: $id) {
    average_rating
    description
    duration
    featured_image
    genre
    id
    crews {
      cast {
        first_name
        last_name
        id
        photo_url
      }
    }
    movie_thumbnails {
      image_url
    }
    my_director {
      first_name
      id
      last_name
      photo_url
    }
    published_at
    ratings {
      feedback
      rating
      userByUser {
        first_name
        last_name
        id
      }
    }
    title
    total_rating
    schedules {
      format
      hall
      id
      movie
      price
      start_time
    }
  }
}
`

export const MOVIES_QUERY = gql`
query MyQuery($where: movies_bool_exp, $offset: Int!) {
  movies(where: $where, limit: 10, offset: $offset) {
    average_rating
    duration
    featured_image
    genre
    id
    published_at
    title
  }
  movies_aggregate(where: $where) {
    aggregate {
      count
    }
  }
}
` 