package service

import (
	"service_orders/internal/model"
)

type OrderService struct {
	repo OrderRepository
}

func NewOrderService(r OrderRepository) *OrderService {
	return &OrderService{repo: r}
}

func (s *OrderService) GetOrder(id int) (*model.Order, error) {
	return s.repo.GetByID(id)
}

func (s *OrderService) ListOrders(userID *int) ([]model.Order, error) {
	orders, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	if userID == nil {
		return orders, nil
	}

	filtered := make([]model.Order, 0)
	for _, o := range orders {
		if o.UserId == *userID {
			filtered = append(filtered, o)
		}
	}

	return filtered, nil
}

func (s *OrderService) CreateOrder(req model.CreateOrderRequest) (int, error) {
	if req.Name == "" || req.Status == "" || req.UserId == 0 {
		return 0, model.ErrMissingRequiredFields
	}
	if req.Price < 0 {
		return 0, model.ErrInvalidPrice
	}

	return s.repo.Create(&req)
}

func (s *OrderService) UpdateOrder(req model.UpdateOrderRequest) error {
	if req.ID == 0 || req.Name == "" || req.Status == "" {
		return model.ErrMissingRequiredFields
	}
	if req.Price < 0 {
		return model.ErrInvalidPrice
	}

	return s.repo.Update(&req)
}

func (s *OrderService) DeleteOrder(id int) error {
	return s.repo.Delete(id)
}