package dto

import (
	"time"

	"github.com/pasokatazip/backend/internal/usecases"
)

type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r CreateUserRequest) ToUseCaseInput() usecases.CreateUserInput {
	return usecases.CreateUserInput{
		Email:    r.Email,
		Password: r.Password,
	}
}

type CreateUserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Subsc     bool      `json:"subsc"`
	CreatedAt time.Time `json:"createdAt"`
}

func NewCreateUserResponse(output usecases.CreateUserOutput) CreateUserResponse {
	return CreateUserResponse{
		ID:        output.ID,
		Email:     output.Email,
		Subsc:     output.Subsc,
		CreatedAt: output.CreatedAt,
	}
}
