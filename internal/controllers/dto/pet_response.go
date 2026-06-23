package dto

import (
	"time"

	"github.com/pasokatazip/backend/internal/usecases"
)

type PetResponse struct {
	ID                   string    `json:"id"`
	Name                 string    `json:"name"`
	IsDeleted            bool      `json:"is_deleted"`
	UserID               string    `json:"user_id"`
	Energy               int       `json:"energy"`
	Curiosity            int       `json:"curiosity"`
	Sociality            int       `json:"sociality"`
	Routine              int       `json:"routine"`
	CurrentGroupMasterID *int      `json:"current_group_master_id"`
	CurrentStageID       int       `json:"current_stage_id"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

func NewPetResponse(output usecases.PetOutput) PetResponse {
	return PetResponse{
		ID:                   output.ID,
		Name:                 output.Name,
		IsDeleted:            output.IsDeleted,
		UserID:               output.UserID,
		Energy:               output.Energy,
		Curiosity:            output.Curiosity,
		Sociality:            output.Sociality,
		Routine:              output.Routine,
		CurrentGroupMasterID: output.CurrentGroupMasterID,
		CurrentStageID:       output.CurrentStageID,
		CreatedAt:            output.CreatedAt,
		UpdatedAt:            output.UpdatedAt,
	}
}
