import gql from 'graphql-tag'

export const ADD_SCHEDULE = gql`
mutation MyMutation($object: [schedules_insert_input!]!) {
  insert_schedules(objects: $object) {
    affected_rows
  }
}
`

export const DELETE_SCHEDULE = gql`
mutation MyMutation($id: Int!) {
  delete_schedules_by_pk(id: $id){
    format
    id
    hall
  }
}
`
export const BOOK_TICKET = gql`
mutation MyMutation($object: tickets_insert_input!) {
  insert_tickets_one(object: $object) {
    id
  }
}
`