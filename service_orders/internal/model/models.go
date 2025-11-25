package model

import "time"

type Order struct {
	ID          int       `json:"id"`
	Name        string    `json:"name,omitempty"`
	Description string    `json:"description,omitempty"`
	Price       int       `json:"price,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
