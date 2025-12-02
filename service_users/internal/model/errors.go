package model

import "errors"

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrMissingRequiredFields = errors.New("missing required fields")
	ErrInvalidEmail          = errors.New("email is invalid")
	ErrUniqueEmailConflict   = errors.New("user with this email is already exists")
	ErrInvalidCredentials    = errors.New("invalid email or password")
	ErrInvalidPassword       = errors.New("invalid password")
)

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}