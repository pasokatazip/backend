package domain

import "errors"

var (
	ErrNotFound             = errors.New("resource not found")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrSubscriptionRequired = errors.New("subscription required")
	ErrValidation           = errors.New("validation error")
	ErrAlreadyExists        = errors.New("resource already exists")
	ErrInternal             = errors.New("internal error")
	ErrExternalService      = errors.New("external service error")
)
