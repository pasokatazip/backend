package dto

import (
	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/usecases"
)

type CreatePetRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

func (r CreatePetRequest) ToUseCaseInput(userID domain.UserID) usecases.CreatePetInput {
	return usecases.CreatePetInput{
		Name:   r.Name,
		Color:  r.Color,
		UserID: userID,
	}
}

type CreatePetResponse = PetResponse

func NewCreatePetResponse(output usecases.PetOutput) CreatePetResponse {
	return NewPetResponse(output)
}
