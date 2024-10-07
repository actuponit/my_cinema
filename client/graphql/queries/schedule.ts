
export const SCHEDULES_QURY = gql`
query MyQuery($where: schedules_bool_exp, $offset: Int!) {
  schedules_aggregate(where: $where) {
    aggregate {
      count
    }
  }
  schedules(limit: 3, offset: $offset, where: $where, order_by: {start_time: desc_nulls_last}) {
    format
    hall
    id
    price
    start_time
    movieByMovie {
      title
    }
  }
}
`

export const SCHEDULES_QURY_BYID = gql`
query MyQuery($id: Int!) {
  schedules_by_pk(id: $id) {
    format
    hall
    movieByMovie {
      title
    }
    price
    start_time
    tickets {
      quantity
      created_at
      user {
        email
        first_name
        last_name
      }
    }
  }
}
`

export const TICKETS_QUERY = gql`
query MyQuery($_eq: Int!) {
  tickets(where: {schedule_id: {_eq: $_eq}}) {
    created_at
    id
    quantity
    updated_at
    user {
      email
      first_name
      last_name
    }
  }
}
`