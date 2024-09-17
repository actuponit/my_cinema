package repositories

import (
	"cinema-server/domain"
	"errors"
)

type UserRepositoryInterface interface {
	CreateUser(user domain.User) (domain.User, error)
	GetUser(email string) (domain.User, error)
}

type UserRepository struct {
	users []domain.User
}

func NewUserRepository(users []domain.User) *UserRepository {
	return &UserRepository{
		users: users,
	}
}

func (r *UserRepository) CreateUser(user domain.User) (domain.User, error) {
	r.users = append(r.users, user)
	return user, nil
}

func (r *UserRepository) GetUser(email string) (domain.User, error) {
	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}
	return domain.User{}, errors.New("user not found")
}
