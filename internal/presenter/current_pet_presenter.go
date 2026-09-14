package presenter

import (
	"github.com/pasokatazip/backend/internal/controllers/dto"
	"github.com/pasokatazip/backend/internal/usecases"
)

type CurrentPetPresenter struct{}

func NewCurrentPetPresenter() *CurrentPetPresenter {
	return &CurrentPetPresenter{}
}

func (p *CurrentPetPresenter) Output(output usecases.FindMyActivePetOutput) dto.CurrentPetResponse {
	response := dto.CurrentPetResponse{
		ID:             output.ID,
		Name:           output.Name,
		Color:          output.Color,
		CurrentStageID: output.CurrentStageID,
		CreatedAt:      output.CreatedAt,
		UpdatedAt:      output.UpdatedAt,
	}
	if output.CurrentGroup != nil {
		response.CurrentGroup = &dto.CurrentGroupResponse{
			ID:          output.CurrentGroup.ID,
			GroupKey:    output.CurrentGroup.GroupKey,
			DisplayName: output.CurrentGroup.DisplayName,
		}
	}
	if output.Departure != nil {
		response.Departure = &dto.DepartureResponse{
			Status:               output.Departure.Status,
			EligibleAt:           output.Departure.EligibleAt,
			ScheduledDepartureAt: output.Departure.ScheduledDepartureAt,
			CanDepart:            output.Departure.CanDepart,
		}
	}
	return response
}
