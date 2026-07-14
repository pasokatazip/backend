package dto

import "github.com/pasokatazip/backend/internal/usecases"

type CurrentGroupResponse struct {
	ID          int    `json:"id"`
	GroupKey    string `json:"group_key"`
	DisplayName string `json:"display_name"`
}

type CurrentPetResponse struct {
	ID           string                `json:"id"`
	Name         string                `json:"name"`
	Color        string                `json:"color"`
	CurrentGroup *CurrentGroupResponse `json:"current_group"`
}

func NewCurrentPetResponse(output usecases.FindMyActivePetOutput) CurrentPetResponse {
	response := CurrentPetResponse{
		ID:    output.ID,
		Name:  output.Name,
		Color: output.Color,
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
