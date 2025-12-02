package model

import "time"

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email,omitempty"`
	Name      string    `json:"name,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Order struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	UserId      int       `json:"userId"`
	Status      string    `json:"status"`
	Price       int       `json:"price"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type UserDetailsResponse struct {
	User   User   `json:"user"`
	Orders []Order `json:"orders"`
}
