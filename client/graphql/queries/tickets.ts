import gql from 'graphql-tag'

export const TICKET_QUERY = gql`
query MyQuery {
  tickets {
    created_at
    id
    quantity
    schedule {
      format
      hall
      start_time
      price
      movieByMovie {
        title
        featured_image
      }
    }
  }
}
`