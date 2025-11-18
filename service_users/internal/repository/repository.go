package repository

import (
	"service_users/internal/model"
	"sync"
)

type UserRepository struct {
	users []model.User
	mu sync.RWMutex
}

func NewCourierRepository() *UserRepository {
	return &UserRepository{
		users: make([]model.User, 0),
	}
}