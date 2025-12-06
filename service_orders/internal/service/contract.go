package service

import (
	"context"
	"service_orders/internal/model"
)

type OrderRepository interface {
	GetByID(id int) (*model.Order, error)
	GetAll() ([]model.Order, error)
	Create(req *model.CreateOrderRequest) (int, error)
	Update(req *model.UpdateOrderRequest) error
	Delete(id int) error
}

type UserChecker interface {
	UserExists(ctx context.Context, userID int) (bool, error)
}