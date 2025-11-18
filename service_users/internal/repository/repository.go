package repository

import (
	"service_users/internal/model"
	"sync"
)

type UserRepository struct {
	mu      sync.RWMutex
	storage map[int]model.User
	nextID  int
}

func NewCourierRepository() *UserRepository {
	return &UserRepository{
		storage: make(map[int]model.User),
		nextID:  1,
	}
}

func (r *UserRepository) GetAll() ([]model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	res := make([]model.User, 0, len(r.storage))
	for _, u := range r.storage {
		res = append(res, u)
	}
	return res, nil
}

func (r *UserRepository) GetByID(id int) (model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.storage[id]
	if !ok {
		return model.User{}, model.ErrUserNotFound
	}
	return u, nil
}

func (r *UserRepository) Create(u model.User) (model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	u.ID = r.nextID
	r.nextID++

	r.storage[u.ID] = u
	return u, nil
}

func (r *UserRepository) Update(u model.User) (model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.storage[u.ID]; !ok {
		return model.User{}, model.ErrUserNotFound
	}

	r.storage[u.ID] = u
	return u, nil
}

func (r *UserRepository) Delete(id int) (model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	u, ok := r.storage[id]
	if !ok {
		return model.User{}, model.ErrUserNotFound
	}

	delete(r.storage, id)
	return u, nil
}
