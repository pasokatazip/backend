package dto

import (
	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/usecases"
)

// UpdatePetDepartureStatusRequest accepts only valid, forward departure states.
type UpdatePetDepartureStatusRequest struct {
	Status string `json:"status"`
}

func (r UpdatePetDepartureStatusRequest) ToUseCaseInput(userID domain.UserID) usecases.UpdatePetDepartureStatusInput {
	return usecases.UpdatePetDepartureStatusInput{
		UserID: userID,
		Status: r.Status,
	}
}
