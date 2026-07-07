package dto

import (
	"github.com/pasokatazip/backend/internal/usecases"
)

type UpdateUserEmailRequest struct {
	CurrentEmail    string `json:"current_email"`
	CurrentPassword string `json:"current_password"`
	NewEmail        string `json:"new_email"`
}

type UpdateUserResponse struct {
	Message string `json:"message"`
}

func (r UpdateUserEmailRequest) ToUseCaseInput() usecases.UpdateUserEmailInput {
	return usecases.UpdateUserEmailInput{
		CurrentEmail:    r.CurrentEmail,
		CurrentPassword: r.CurrentPassword,
		NewEmail:        r.NewEmail,
	}
}
