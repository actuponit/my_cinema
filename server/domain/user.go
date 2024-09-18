package domain

type User struct {
	ID        int    `json:"id" graphql:"id"`
	FirstName string `json:"first_name" graphql:"first_name"`
	LastName  string `json:"last_name" graphql:"last_name"`
	Email     string `json:"email" graphql:"email"`
	Password  string `json:"password" graphql:"password"`
	Role      string `json:"role" default:"user" graphql:"role"`
}
