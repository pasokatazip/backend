package presenter

import (
	"github.com/pasokatazip/backend/internal/controllers/dto"
	"github.com/pasokatazip/backend/internal/usecases"
)

type UpdatePetDepartureStatusPresenter struct{}

type HistoricalReportPresenter struct{}

type RunHourlySimulationPresenter struct{}

func NewUpdatePetDepartureStatusPresenter() *UpdatePetDepartureStatusPresenter {
	return &UpdatePetDepartureStatusPresenter{}
}

func NewHistoricalReportPresenter() *HistoricalReportPresenter {
	return &HistoricalReportPresenter{}
}

func NewRunHourlySimulationPresenter() *RunHourlySimulationPresenter {
	return &RunHourlySimulationPresenter{}
}

func (p *UpdatePetDepartureStatusPresenter) Output(output usecases.UpdatePetDepartureStatusOutput) dto.UpdatePetDepartureStatusResponse {
	return dto.UpdatePetDepartureStatusResponse{
		PetID:                output.PetID,
		Status:               output.Status,
		EligibleAt:           output.EligibleAt,
		ScheduledDepartureAt: output.ScheduledDepartureAt,
		DepartedAt:           output.DepartedAt,
	}
}

func (p *HistoricalReportPresenter) Output(outputs []usecases.FindByTodayReportOutput) []dto.HistoricalReportResponse {
	if outputs == nil {
		return nil
	}

	responses := make([]dto.HistoricalReportResponse, 0, len(outputs))
	for _, output := range outputs {
		responses = append(responses, dto.HistoricalReportResponse{
			ID:        output.ID,
			PetID:     string(output.PetID),
			HourSlot:  output.HourSlot,
			Gossip:    output.Gossip,
			GroupName: output.GroupName,
			CreatedAt: output.CreatedAt,
			Rumors:    output.Rumors,
		})
	}
	return responses
}

func (p *RunHourlySimulationPresenter) Output(output usecases.RunHourlyPetSimulationOutput) dto.RunHourlySimulationResponse {
	var results []dto.RunHourlySimulationPetResultResponse
	if output.Results != nil {
		results = make([]dto.RunHourlySimulationPetResultResponse, 0, len(output.Results))
		for _, result := range output.Results {
			results = append(results, dto.RunHourlySimulationPetResultResponse{
				PetID:           result.PetID,
				PreviousGroupID: result.PreviousGroupID,
				NextGroupID:     result.NextGroupID,
				Moved:           result.Moved,
				MoveProbability: result.MoveProbability,
				AmbientEvent:    result.AmbientEvent,
			})
		}
	}

	return dto.RunHourlySimulationResponse{
		SimulatedAt:          output.SimulatedAt,
		TotalPets:            output.TotalPets,
		Processed:            output.Processed,
		Skipped:              output.Skipped,
		InterestPropagations: output.InterestPropagations,
		ReportsCreated:       output.ReportsCreated,
		Results:              results,
	}
}
