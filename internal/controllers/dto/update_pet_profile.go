package dto

import (
	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/usecases"
)

type UpdatePetProfileRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

func (r UpdatePetProfileRequest) ToUseCaseInput(
	petID domain.PetID,
	userID domain.UserID,
) usecases.UpdatePetProfileInput {
	return usecases.UpdatePetProfileInput{
		PetID:  petID,
		UserID: userID,
		Name:   r.Name,
		Color:  r.Color,
	}
}

func NewUpdatePetProfileResponse(output usecases.PetOutput) PetResponse {
	return NewPetResponse(output)
}
