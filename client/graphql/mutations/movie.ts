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

export const MOVIE_UPDATE = gql`
mutation MyMutation($id: Int!, $_set: movies_set_input!) {
  update_movies_by_pk(pk_columns: {id: $id}, _set: $_set) {
    id
    director
    genre
    duration
    featured_image
    description
    published_at
    title
  }
}
`

export const MOVIE_THUMBNAIL_DELETE = gql`
mutation MyMutation($image_url: String!) {
  delete_movie_thumbnails_by_pk(image_url: $image_url) {
    image_url
    movie_id
  }
}
`

export const MOVIE_THUMBNAIL_ADD = gql`
mutation MyMutation($objects: [movie_thumbnails_insert_input!]!) {
  insert_movie_thumbnails(objects: $objects) {
    affected_rows
  }
}
`

export const BOOKMARK_MOVIE = gql`
mutation MyMutation($object: bookmarks_insert_input!) {
  insert_bookmarks_one(object: $object) {
    movie_id
    user
  }
}
`

export const DELETE_BOOKMARK_MUTATION = gql`
mutation MyMutation($movie_id: Int!, $user: Int!) {
  delete_bookmarks_by_pk(movie_id: $movie_id, user: $user) {
    movie_id
  }
}
`

export const EDIT_REVIEW = gql`
mutation MyMutation($id: Int!, $set: ratings_set_input) {
  update_ratings(where: {movie: {_eq: $id}}, _set: $set) {
    affected_rows
  }
}
`