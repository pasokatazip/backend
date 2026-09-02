package usecases

import (
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type FindActivePetEvolutionHistoryInput struct {
	UserID domain.UserID
}

type FindActivePetEvolutionHistoryOutput struct {
	PetID          string                          `json:"pet_id"`
	CreatedAt      time.Time                       `json:"created_at"`
	CurrentStageID int                             `json:"current_stage_id"`
	Stages         []ActivePetEvolutionStageOutput `json:"stages"`
	Evolutions     []PetGrowthEvolutionOutput      `json:"evolutions"`
}

type ActivePetEvolutionStageOutput struct {
	ID        int        `json:"id"`
	StageKey  string     `json:"stage_key"`
	StageNo   int        `json:"stage_no"`
	Name      string     `json:"name"`
	BranchKey *string    `json:"branch_key,omitempty"`
	ImageURL  *string    `json:"image_url,omitempty"`
	Unlocked  bool       `json:"unlocked"`
	Current   bool       `json:"current"`
	EvolvedAt *time.Time `json:"evolved_at,omitempty"`
}

type FindActivePetEvolutionHistory struct {
	petRepo       domain.PetRepository
	stageRepo     domain.EvolutionStageRepository
	evolutionRepo domain.PetEvolutionRepository
}

func NewFindActivePetEvolutionHistory(
	petRepo domain.PetRepository,
	stageRepo domain.EvolutionStageRepository,
	evolutionRepo domain.PetEvolutionRepository,
) *FindActivePetEvolutionHistory {
	return &FindActivePetEvolutionHistory{
		petRepo:       petRepo,
		stageRepo:     stageRepo,
		evolutionRepo: evolutionRepo,
	}
}

func (u *FindActivePetEvolutionHistory) Execute(
	input FindActivePetEvolutionHistoryInput,
) (FindActivePetEvolutionHistoryOutput, error) {
	if !domain.IsValidUserID(input.UserID) {
		return FindActivePetEvolutionHistoryOutput{}, domain.ErrValidation
	}

	pet, err := u.petRepo.FindActiveByUserID(input.UserID)
	if err != nil {
		return FindActivePetEvolutionHistoryOutput{}, err
	}

	stages, err := u.stageRepo.FindAll()
	if err != nil {
		return FindActivePetEvolutionHistoryOutput{}, err
	}

	evolutions, err := u.evolutionRepo.FindByPetID(pet.ID())
	if err != nil {
		return FindActivePetEvolutionHistoryOutput{}, err
	}

	return FindActivePetEvolutionHistoryOutput{
		PetID:          string(pet.ID()),
		CreatedAt:      pet.CreatedAt(),
		CurrentStageID: pet.CurrentStageID(),
		Stages:         newActivePetEvolutionStageOutputs(stages, evolutions, pet.CurrentStageID()),
		Evolutions:     newPetGrowthEvolutionOutputs(evolutions),
	}, nil
}

func newActivePetEvolutionStageOutputs(
	stages []domain.EvolutionStage,
	evolutions []domain.PetEvolution,
	currentStageID int,
) []ActivePetEvolutionStageOutput {
	evolvedAtByStageID := make(map[int]time.Time, len(evolutions))
	unlockedStageIDs := map[int]bool{0: true}
	for _, evolution := range evolutions {
		stageID := int(evolution.StageID())
		evolvedAtByStageID[stageID] = evolution.EvolvedAt()
		unlockedStageIDs[stageID] = true
	}
	unlockedStageIDs[currentStageID] = true

	outputs := make([]ActivePetEvolutionStageOutput, 0, len(stages))
	for _, stage := range stages {
		stageID := int(stage.ID())
		evolvedAt, evolved := evolvedAtByStageID[int(stage.ID())]
		outputs = append(outputs, ActivePetEvolutionStageOutput{
			ID:        stageID,
			StageKey:  stage.StageKey(),
			StageNo:   stage.StageNo(),
			Name:      stage.Name(),
			BranchKey: stage.BranchKey(),
			ImageURL:  stage.ImageURL(),
			Unlocked:  unlockedStageIDs[stageID],
			Current:   stageID == currentStageID,
			EvolvedAt: nullableEvolvedAt(evolvedAt, evolved),
		})
	}
	return outputs
}

func nullableEvolvedAt(evolvedAt time.Time, evolved bool) *time.Time {
	if !evolved {
		return nil
	}
	return &evolvedAt
}
