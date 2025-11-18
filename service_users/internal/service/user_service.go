package service

import (
	"regexp"
	"service_users/internal/model"
)

type UserService struct {
	repository UserRepository
}

func NewUserService(r UserRepository) *UserService {
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

func (s *UserService) UpdateUser(req model.UpdateUserRequest) error {
	if req.Name == "" {
		return model.ErrMissingRequiredFields
	}

	_, err := s.repository.GetByID(req.ID)
	if err != nil {
		return err
	}

	return s.repository.Update(&req)
}

func (s *UserService) DeleteUser(id int) error {
	return s.repository.Delete(id)
}

// Helpers

func isEmailValid(e string) bool {
    emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
    return emailRegex.MatchString(e)
}