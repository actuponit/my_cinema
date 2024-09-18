package repositories

import (
	"cinema-server/domain"
	"context"
	"errors"
	"log"

	"github.com/hasura/go-graphql-client"
)

type UserRepositoryInterface interface {
	CreateUser(user domain.User) (domain.User, error)
	GetUser(email string) (domain.User, error)
}

type UserRepository struct {
	client *graphql.Client
}

func NewUserRepository(client *graphql.Client) *UserRepository {
	return &UserRepository{
		client: client,
	}
}

func (r *UserRepository) CreateUser(user domain.User) (domain.User, error) {
	var m struct {
		InsertUser domain.User `graphql:"insert_users_one(object: {email: $email, password: $password, first_name: $first_name, last_name: $last_name})"`
	}

	log.Println(user)
	variables := map[string]interface{}{
		"email":      user.Email,
		"password":   user.Password,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
	}

	err := r.client.Mutate(context.Background(), &m, variables)

	if err != nil {
		return domain.User{}, err
	}

	return m.InsertUser, nil
}

func (r *UserRepository) GetUser(email string) (domain.User, error) {
	var q struct {
		Users []domain.User `graphql:"users(where: {email: {_eq: $email}})"`
	}
	variables := map[string]interface{}{
		"email": email,
	}
	err := r.client.Query(context.Background(), &q, variables)
	if err != nil {
		return domain.User{}, err
	}
	log.Println(q.Users)
	if len(q.Users) == 0 {
		return domain.User{}, errors.New("user not found")
	}

	return q.Users[0], nil
}
