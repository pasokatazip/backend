package domain

import "errors"

var (
	ErrNotFound      = errors.New("user not found")
	
	ErrUnauthorized = errors.New("unauthorized")
	ErrSubscriptionRequired = errors.New("subscription required")


	ErrValidation = errors.New("validation error")
	ErrAlreadyExists = errors.New("resource already exists")
)