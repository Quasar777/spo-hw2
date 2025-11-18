package service

import "service_users/internal/model"

type UserRepository interface {
	GetAll() ([]model.User, error)
	GetByID(id int) (model.User, error)
	Create(u model.User) (model.User, error)
	Update(u model.User) (model.User, error)
	Delete(id int) (model.User, error) 
}