package service

import "service_users/internal/model"

type UserRepository interface {
	GetByID(id int) (*model.User, error)
	GetAll() ([]model.User, error)
	Create(req *model.CreateUserRequest) (int, error)
	Update(req *model.UpdateUserRequest) error
	Delete(id int) error 
}