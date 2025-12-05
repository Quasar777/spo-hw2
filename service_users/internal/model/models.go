package model

import "time"

type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email,omitempty"`
	Name         string    `json:"name,omitempty"`
	PasswordHash string    `json:"-"`     // не отдаём наружу
	Roles        []string  `json:"roles"` // например ["user"], ["admin"]
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type CreateUserRequest struct {
	Email        string `json:"email,omitempty"`
	Name         string `json:"name,omitempty"`
	PasswordHash string
	Roles        []string
}

type UpdateUserRequest struct {
	ID   int    `json:"id"`
	Name string `json:"name,omitempty"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateProfileRequest struct {
	Name string `json:"name"`
}
