package model

import "time"

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email,omitempty"`
	Name      string    `json:"name,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CreateUserRequest struct {
	Email     string    `json:"email,omitempty"`
	Name      string    `json:"name,omitempty"`
}

type UpdateUserRequest struct {
	ID        int       `json:"id"`
	Name      string    `json:"name,omitempty"`	
}