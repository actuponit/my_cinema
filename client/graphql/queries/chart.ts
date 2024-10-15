import gql from 'graphql-tag'

export const CHART_QUERY = gql`
  query MyQuery {
  tickets_graph {
    count
    t_date
    total_price
  }
}
` 
export const CHART_GENRE_QUERY = gql`
  query MyQuery {
    group_movie_by_category {
      count
      genre
    }
  }
` 
