package presenter

import (
	"github.com/pasokatazip/backend/internal/controllers/dto"
	"github.com/pasokatazip/backend/internal/usecases"
)

type ActivePetEvolutionHistoryPresenter struct{}
type CurrentPetEvolutionStatusPresenter struct{}

func NewActivePetEvolutionHistoryPresenter() *ActivePetEvolutionHistoryPresenter { return &ActivePetEvolutionHistoryPresenter{} }
func NewCurrentPetEvolutionStatusPresenter() *CurrentPetEvolutionStatusPresenter { return &CurrentPetEvolutionStatusPresenter{} }

func (p *ActivePetEvolutionHistoryPresenter) Output(output usecases.FindActivePetEvolutionHistoryOutput) dto.ActivePetEvolutionHistoryResponse {
	return dto.ActivePetEvolutionHistoryResponse{
		PetID: output.PetID, CreatedAt: output.CreatedAt, CurrentStageKey: output.CurrentStageKey,
		Stages: evolutionStageResponses(output.Stages), Evolutions: petGrowthEvolutionResponses(output.Evolutions),
	}
}

func (p *CurrentPetEvolutionStatusPresenter) Output(output usecases.FindCurrentPetEvolutionStatusOutput) dto.CurrentPetEvolutionStatusResponse {
	var candidates []dto.EvolutionCandidateResponse
	if output.NextStages != nil {
		candidates = make([]dto.EvolutionCandidateResponse, 0, len(output.NextStages))
		for _, candidate := range output.NextStages {
			candidates = append(candidates, dto.EvolutionCandidateResponse{
				RuleID: candidate.RuleID, ToStage: currentEvolutionStageResponse(candidate.ToStage),
				SelectedForPet: candidate.SelectedForPet, Ready: candidate.Ready,
				Requirements: evolutionRequirementsResponse(candidate.Requirements),
			})
		}
	}
	return dto.CurrentPetEvolutionStatusResponse{
		PetID: output.PetID, CurrentStage: currentEvolutionStageResponse(output.CurrentStage),
		CanEvolve: output.CanEvolve, NextStages: candidates,
	}
}

func evolutionStageResponses(outputs []usecases.ActivePetEvolutionStageOutput) []dto.EvolutionStageResponse {
	if outputs == nil { return nil }
	responses := make([]dto.EvolutionStageResponse, 0, len(outputs))
	for _, output := range outputs {
		responses = append(responses, dto.EvolutionStageResponse{
			ID: output.ID, StageKey: output.StageKey, StageNo: output.StageNo, Name: output.Name,
			BranchKey: output.BranchKey, ImageURL: output.ImageURL, Unlocked: output.Unlocked,
			Current: output.Current, EvolvedAt: output.EvolvedAt,
		})
	}
	return responses
}

func petGrowthEvolutionResponses(outputs []usecases.PetGrowthEvolutionOutput) []dto.PetGrowthEvolutionResponse {
	if outputs == nil { return nil }
	responses := make([]dto.PetGrowthEvolutionResponse, 0, len(outputs))
	for _, output := range outputs {
		responses = append(responses, dto.PetGrowthEvolutionResponse{
			ID: output.ID, StageID: output.StageID, EvolutionRuleID: output.EvolutionRuleID,
			PrimaryStatus: output.PrimaryStatus, EvolvedAt: output.EvolvedAt, CreatedAt: output.CreatedAt,
		})
	}
	return responses
}

func currentEvolutionStageResponse(output usecases.CurrentPetEvolutionStageOutput) dto.CurrentEvolutionStageResponse {
	return dto.CurrentEvolutionStageResponse{ID: output.ID, StageKey: output.StageKey, StageNo: output.StageNo, Name: output.Name, BranchKey: output.BranchKey, ImageURL: output.ImageURL}
}
func evolutionProgressResponse(output usecases.CurrentPetEvolutionProgress) dto.EvolutionProgressResponse {
	return dto.EvolutionProgressResponse{Current: output.Current, Required: output.Required, Remaining: output.Remaining, Met: output.Met}
}
func evolutionRequirementsResponse(output usecases.CurrentPetEvolutionRequirements) dto.EvolutionRequirementsResponse {
	return dto.EvolutionRequirementsResponse{Experience: evolutionProgressResponse(output.Experience), FeedCount: evolutionProgressResponse(output.FeedCount), DaysSinceLastEvolution: evolutionProgressResponse(output.DaysSinceLastEvolution)}
}
