package dto

import (
	"time"

	"github.com/pasokatazip/backend/internal/usecases"
)

type PetResponse struct {
	ID                   string    `json:"id"`
	Name                 string    `json:"name"`
	Color                string    `json:"color"`
	IsDeleted            bool      `json:"is_deleted"`
	UserID               string    `json:"user_id"`
	Energy               float64   `json:"energy"`
	Curiosity            float64   `json:"curiosity"`
	Sociality            float64   `json:"sociality"`
	Routine              float64   `json:"routine"`
	CurrentGroupMasterID *int      `json:"current_group_master_id"`
	CurrentStageID       int       `json:"current_stage_id"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type PetListResponse struct {
	Pets []PetResponse `json:"pets"`
}

type AllPetResponse struct {
	PetID string `json:"pet_id"`
	Name  string `json:"name"`
}

type AllPetListResponse struct {
	Pets []AllPetResponse `json:"pets"`
}

func NewPetResponse(output usecases.PetOutput) PetResponse {
	return PetResponse{
		ID:                   output.ID,
		Name:                 output.Name,
		Color:                output.Color,
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

func NewPetListResponse(outputs []usecases.PetOutput) PetListResponse {
	pets := make([]PetResponse, 0, len(outputs))
	for _, output := range outputs {
		pets = append(pets, NewPetResponse(output))
	}

	return PetListResponse{Pets: pets}
}

func NewAllPetListResponse(outputs []usecases.PetOutput) AllPetListResponse {
	pets := make([]AllPetResponse, 0, len(outputs))
	for _, output := range outputs {
		pets = append(pets, AllPetResponse{
			PetID: output.ID,
			Name:  output.Name,
		})
	}

	return AllPetListResponse{Pets: pets}
}
