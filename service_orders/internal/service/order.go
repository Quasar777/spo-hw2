package service

import (
	"context"
	"fmt"
	"service_orders/internal/model"
)

type OrderService struct {
	repo OrderRepository
	userChecker UserChecker
}

func NewOrderService(r OrderRepository, uc UserChecker) *OrderService {
	return &OrderService{repo: r, userChecker: uc}
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

func (s *OrderService) CreateOrder(ctx context.Context, req model.CreateOrderRequest) (int, error) {
	if req.Name == "" || req.Status == "" || req.UserId == 0 {
		return 0, model.ErrMissingRequiredFields
	}
	if req.Price < 0 {
		return 0, model.ErrInvalidPrice
	}

	exists, err := s.userChecker.UserExists(ctx, req.UserId)
	if err != nil {
		return 0, fmt.Errorf("user check failed: %w", err)
	}
	if !exists {
		return 0, model.ErrUserNotFound
	}

	return s.repo.Create(&req)
}

func (s *OrderService) UpdateOrder(req model.UpdateOrderRequest) error {
	if req.Name == "" && req.Status == "" && req.Description == "" {
		return model.ErrMissingRequiredFields
	}
	if req.Price < 0 {
		return model.ErrInvalidPrice
	}

	existingOrder, err := s.repo.GetByID(req.ID)
	if err != nil {
		return err
	}

	if req.Name == "" { req.Name = existingOrder.Name }
	if req.Status == "" { req.Status = existingOrder.Status }
	if req.Description == "" { req.Description = existingOrder.Description }

	return s.repo.Update(&req)
}

func (s *OrderService) DeleteOrder(id int) error {
	return s.repo.Delete(id)
}