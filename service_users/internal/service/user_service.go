package service

import (
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

