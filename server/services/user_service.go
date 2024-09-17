package services

import (
	"cinema-server/domain"
	"cinema-server/repositories"
	"cinema-server/utils"
)

type UserServiceInterface interface {
	SignUp(user domain.User) (domain.UserDto, error)
	Login(email string, password string) (domain.UserDto, error)
}

type UserService struct {
	repo repositories.UserRepositoryInterface
}

func NewUserService(repo repositories.UserRepositoryInterface) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) SignUp(user domain.User) (domain.UserDto, error) {
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return domain.UserDto{}, err
	}
	user.Password = string(hashedPassword)
	user, err = s.repo.CreateUser(user)
	if err != nil {
		return domain.UserDto{}, err
	}
	accessToken, err := utils.GenerateToken(user.ID, user.Role)
	if err != nil {
		return domain.UserDto{}, err
	}
	return domain.UserDto{User: user, AccessToken: accessToken}, nil
}

func (s *UserService) Login(email string, password string) (domain.UserDto, error) {
	user, err := s.repo.GetUser(email)
	if err != nil {
		return domain.UserDto{}, err
	}
	err = utils.IsPasswordCorrect(password, user.Password)
	if err != nil {
		return domain.UserDto{}, err
	}
	accessToken, err := utils.GenerateToken(user.ID, user.Role)
	if err != nil {
		return domain.UserDto{}, err
	}

	return domain.UserDto{User: user, AccessToken: accessToken}, nil
}
