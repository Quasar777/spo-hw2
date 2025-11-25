package service

import "service_orders/internal/model"

type OrderRepository interface {
	GetByID(id int) (model.Order, error)
	GetAll() ([]model.Order, error)
	Create(payload map[string]any) (model.Order, error)
	Update(id int, updates map[string]any) (model.Order, error)
	Delete(id int) (model.Order, error)
}
