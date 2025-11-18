package service

import (
	"regexp"
	"service_users/internal/model"
	"service_users/internal/repository"
)

type UserService struct {
	repository *repository.UserRepository
}

func NewUserService(r *repository.UserRepository) *UserService {
	return &UserService{repository: r}
}

func (s *UserService) GetUser(id int) (*model.User, error) {
	return s.repository.GetByID(id)
}

func (s *UserService) GetAllUsers() ([]model.User, error) {
	return s.repository.GetAll()
}

func (s *UserService) CreateUser(req model.CreateUserRequest) (int, error) {
	if req.Name == "" || req.Email == "" {
		return 0, model.ErrMissingRequiredFields
	}
	if !isEmailValid(req.Email) {
		return 0, model.ErrInvalidEmail
	}

	return s.repository.Create(&req)
}


// Helpers

func isEmailValid(e string) bool {
    emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
    return emailRegex.MatchString(e)
}