package dto

import (
	"time"

	"github.com/pasokatazip/backend/internal/usecases"
)

type CurrentGroupResponse struct {
	ID          int    `json:"id"`
	GroupKey    string `json:"group_key"`
	DisplayName string `json:"display_name"`
}

type CurrentPetResponse struct {
	ID             string                `json:"id"`
	Name           string                `json:"name"`
	Color          string                `json:"color"`
	CurrentStageID int                   `json:"current_stage_id"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`
	CurrentGroup   *CurrentGroupResponse `json:"current_group"`
}

func NewCurrentPetResponse(output usecases.FindMyActivePetOutput) CurrentPetResponse {
	response := CurrentPetResponse{
		ID:             output.ID,
		Name:           output.Name,
		Color:          output.Color,
		CurrentStageID: output.CurrentStageID,
		CreatedAt:      output.CreatedAt,
		UpdatedAt:      output.UpdatedAt,
	}
	if output.CurrentGroup != nil {
		response.CurrentGroup = &CurrentGroupResponse{
			ID:          output.CurrentGroup.ID,
			GroupKey:    output.CurrentGroup.GroupKey,
			DisplayName: output.CurrentGroup.DisplayName,
		}
	}
	return response
}
