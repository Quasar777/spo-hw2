package repository

import (
	"service_orders/internal/model"
	"sync"
	"time"
)

type InMemoryOrderRepository struct {
	mu      sync.RWMutex
	storage map[int]model.Order
	nextID  int
}

func NewInMemoryOrderRepository() *InMemoryOrderRepository {
	r := &InMemoryOrderRepository{
		storage: make(map[int]model.Order),
		nextID:  4,
	}

	r.storage[1] = model.Order{
		ID:          1,
		Name:        "Pizza Margherita",
		Description: "Classic pizza with tomatoes and cheese",
		Price:       1200,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	r.storage[2] = model.Order{
		ID:          2,
		Name:        "Burger XXL",
		Description: "Double beef burger with fries",
		Price:       1500,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	r.storage[3] = model.Order{
		ID:          3,
		Name:        "Latte",
		Description: "Coffee latte 400ml",
		Price:       450,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return r
}


func (r *InMemoryOrderRepository) GetByID(id int) (*model.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	order, ok := r.storage[id]
	if !ok {
		return nil, model.ErrOrderNotFound
	}

	// Здесь возвращается копия, чтобы снаружи не мутировали map
	o := order
	return &o, nil
}

func (r *InMemoryOrderRepository) GetAll() ([]model.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	orders := make([]model.Order, 0, len(r.storage))
	for _, o := range r.storage {
		orders = append(orders, o)
	}

	return orders, nil
}

func (r *InMemoryOrderRepository) Create(req *model.CreateOrderRequest) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := r.nextID
	r.nextID++

	order := model.Order{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		UserId:      req.UserId,
		Status:      req.Status,
		Price:       req.Price,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	r.storage[id] = order

	o := order
	return o.ID, nil
}


func (r *InMemoryOrderRepository) Update(req *model.UpdateOrderRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	order, ok := r.storage[req.ID]
	if !ok {
		return model.ErrOrderNotFound
	}

	order.Name = req.Name
	order.Description = req.Description
	order.Price = req.Price
	order.Status = req.Status
	order.UpdatedAt = time.Now()

	r.storage[req.ID] = order

	return nil
}

func (r *InMemoryOrderRepository) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.storage[id]
	if !ok {
		return model.ErrOrderNotFound
	}

	delete(r.storage, id)

	return nil
}