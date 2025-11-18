package model

import "time"

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email,omitempty"`
	Name      string    `json:"name,omitempty"`
	// Можно добавить любые поля, которые ты уже используешь в теле запроса
	// Password и прочее сейчас НЕ трогаем, просто повторяем текущую логику
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type CreateUserRequest struct {
	Email     string    `json:"email,omitempty"`
	Name      string    `json:"name,omitempty"`
}

type UpdateUserRequest struct {
	ID        int       `json:"id"`
	Name      string    `json:"name,omitempty"`	
}