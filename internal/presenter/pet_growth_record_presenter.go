package presenter

import (
	"github.com/pasokatazip/backend/internal/controllers/dto"
	"github.com/pasokatazip/backend/internal/usecases"
)

type PetGrowthRecordPresenter struct{}
func NewPetGrowthRecordPresenter() *PetGrowthRecordPresenter { return &PetGrowthRecordPresenter{} }
func (p *PetGrowthRecordPresenter) Output(output usecases.FindPetGrowthRecordOutput) dto.PetGrowthRecordResponse {
	var events []dto.PetGrowthExperienceEventResponse
	if output.ExperienceEvents != nil {
		events = make([]dto.PetGrowthExperienceEventResponse, 0, len(output.ExperienceEvents))
		for _, event := range output.ExperienceEvents {
			events = append(events, dto.PetGrowthExperienceEventResponse{
				ID: event.ID, SourceType: event.SourceType, SourceID: event.SourceID,
				Amount: event.Amount, CappedAmount: event.CappedAmount,
				ExperienceDate: event.ExperienceDate, CreatedAt: event.CreatedAt,
			})
		}
	}
	return dto.PetGrowthRecordResponse{
		PetID: output.PetID, Color: output.Color, CreatedAt: output.CreatedAt,
		CurrentStageID: output.CurrentStageID, Stages: evolutionStageResponses(output.Stages),
		TotalExperience: output.TotalExperience, FeedCount: output.FeedCount,
		ExperienceEvents: events, Evolutions: petGrowthEvolutionResponses(output.Evolutions),
	}
}
