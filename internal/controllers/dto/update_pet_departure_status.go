package dto

import (
	"time"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/usecases"
)

// UpdatePetDepartureStatusRequest accepts only valid, forward departure states.
type UpdatePetDepartureStatusRequest struct {
	Status string `json:"status"`
}

type UpdatePetDepartureStatusResponse struct {
	PetID                string     `json:"pet_id"`
	Status               string     `json:"status"`
	EligibleAt           time.Time  `json:"eligible_at"`
	ScheduledDepartureAt time.Time  `json:"scheduled_departure_at"`
	DepartedAt           *time.Time `json:"departed_at,omitempty"`
}

func (r UpdatePetDepartureStatusRequest) ToUseCaseInput(userID domain.UserID) usecases.UpdatePetDepartureStatusInput {
	return usecases.UpdatePetDepartureStatusInput{
		UserID: userID,
		Status: r.Status,
	}
}
