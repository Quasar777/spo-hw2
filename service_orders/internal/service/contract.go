package service

import "service_orders/internal/model"

type OrderRepository interface {
	GetByID(id int) (*model.Order, error)
	GetAll() ([]model.Order, error)
	Create(req *model.CreateOrderRequest) (int, error)
	Update(req *model.UpdateOrderRequest) error
	Delete(id int) error
}
