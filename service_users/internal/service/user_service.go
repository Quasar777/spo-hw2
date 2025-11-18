package service

import (
	"service_users/internal/repository"
)

type UserService struct {
	repository *repository.UserRepository
}

func NewCourierUseCase(r *repository.UserRepository) *UserService {
	return &UserService{repository: r}
}