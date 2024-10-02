import gql from 'graphql-tag'

export const MOVIE_INSERT = gql`
mutation MyMutation($object: movies_insert_input!) {
  insert_movies_one(object: $object) {
    average_rating
    director
    id
    genre
    duration
    featured_image
    description
    published_at
    title
    total_rating
  }
}
`

export const MOVIE_REVIEW = gql`
mutation MyMutation($object: ratings_insert_input!) {
  insert_ratings_one(object: $object) {
    feedback
    movie
    user
    rating
  }
}
`