package dto

import (
	"github.com/pasokatazip/backend/internal/usecases"
)

type CreatePetRequest struct {
	Name                 string  `json:"name"`
	UserID               string  `json:"user_id"`
	CurrentGroupMasterID *string `json:"current_group_master_id"`
	CurrentStageID       string  `json:"current_stage_id"`
}

func (r CreatePetRequest) ToUseCaseInput() usecases.CreatePetInput {
	return usecases.CreatePetInput{
		Name:                 r.Name,
		UserID:               r.UserID,
		CurrentGroupMasterID: r.CurrentGroupMasterID,
		CurrentStageID:       r.CurrentStageID,
	}
}

type CreatePetResponse = PetResponse

func NewCreatePetResponse(output usecases.PetOutput) CreatePetResponse {
	return NewPetResponse(output)
}