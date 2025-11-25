package model

import "errors"

var (
	ErrOrderNotFound        = errors.New("order not found")
	ErrMissingRequiredFields = errors.New("missing required fields")
	ErrInvalidPrice         = errors.New("invalid price")
)
