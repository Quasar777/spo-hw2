package model

import "errors"

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrMissingRequiredFields = errors.New("missing required fields")
	ErrInvalidEmail          = errors.New("email is invalid")
	ErrUniqueEmailConflict   = errors.New("user with this email is already exists")
)
