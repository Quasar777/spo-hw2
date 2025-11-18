package service

import "service_users/internal/model"

type UserRepository interface {
	GetByID(id int) (*model.User, error)
	GetAll() ([]model.User, error)
	Create(u *model.CreateUserRequest) (int, error)
	Update(u *model.UpdateUserRequest) error
	Delete(id int) error 
}