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
	r := &UserRepository{
		storage: make(map[int]model.User),
		nextID:  4,
	}

	now := time.Now()

	r.storage[1] = model.User{
		ID:           1,
		Name:         "Alice",
		Email:        "alice@example.com",
		PasswordHash: "",               // временно пусто
		Roles:        []string{"user"}, // по умолчанию обычный пользователь
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	r.storage[2] = model.User{
		ID:           2,
		Name:         "John",
		Email:        "john@example.com",
		PasswordHash: "",
		Roles:        []string{"user"},
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	r.storage[3] = model.User{
		ID:           3,
		Name:         "Andrew",
		Email:        "andrew@example.com",
		PasswordHash: "",
		Roles:        []string{"admin"}, // допустим, ты админ :)
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	return r
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

func (r *UserRepository) Create(req *model.CreateUserRequest) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, u := range r.storage {
		if req.Email == u.Email {
			return 0, model.ErrUniqueEmailConflict
		}
	}

	newUser := &model.User{
		ID:           r.nextID,
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: req.PasswordHash,
		Roles:        req.Roles,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
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
	userDB.UpdatedAt = time.Now()
	r.storage[user.ID] = userDB

	return nil
}


func (r *UserRepository) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.storage[id]
	if !ok {
		return model.ErrUserNotFound
	}

	delete(r.storage, id)
	return nil
}

func (r *UserRepository) GetByEmail(email string) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, u := range r.storage {
		if u.Email == email {
			userCopy := u
			return &userCopy, nil
		}
	}

	return nil, model.ErrUserNotFound
}