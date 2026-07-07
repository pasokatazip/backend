package dto

import (
	"github.com/pasokatazip/backend/internal/usecases"
)

type UpdateUserPasswordRequest struct {
	Email           string `json:"email"`
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (r UpdateUserPasswordRequest) ToUseCaseInput() usecases.UpdateUserPasswordInput {
	return usecases.UpdateUserPasswordInput{
		Email:           r.Email,
		CurrentPassword: r.CurrentPassword,
		NewPassword:     r.NewPassword,
	}
}
