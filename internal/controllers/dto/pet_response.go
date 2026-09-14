package dto

import "time"

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
