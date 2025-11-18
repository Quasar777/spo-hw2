package repository

import (
	"service_users/internal/model"
	"sort"
	"sync"
	"time"
)

type UserRepository struct {
	mu      sync.RWMutex
	storage map[int]model.User
	nextID  int
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		storage: make(map[int]model.User),
		nextID:  1,
	}
}

func (r *UserRepository) GetByID(id int) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.storage[id]
	if !ok {
		return nil, model.ErrUserNotFound
	}
	return &u, nil
}

func (r *UserRepository) GetAll() ([]model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	res := make([]model.User, 0, len(r.storage))
	for _, u := range r.storage {
		res = append(res, u)
	}
	
	sort.Slice(res, func(i, j int) bool {
		return res[i].ID < res[j].ID
	})

	return res, nil
}

func (r *UserRepository) Create(user *model.CreateUserRequest) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, u := range r.storage {
		if user.Email == u.Email {
			return 0, model.ErrUniqueEmailConflict
		}
	}

	newUser := &model.User{
		ID: r.nextID,
		Email: user.Email,
		Name: user.Name,
		CreatedAt: time.Now(),
	}
	r.nextID++
	
	r.storage[newUser.ID] = *newUser
	return newUser.ID, nil
}

func (r *UserRepository) Update(user *model.UpdateUserRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	userDB, ok := r.storage[user.ID]
	if !ok {
		return model.ErrUserNotFound
	}

	for _, u := range r.storage {
		if userDB.Email == u.Email && userDB.ID != u.ID {
			return model.ErrUniqueEmailConflict
		}
	}
	
	userDB.Name = user.Name
	r.storage[user.ID] = userDB

	return nil
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
