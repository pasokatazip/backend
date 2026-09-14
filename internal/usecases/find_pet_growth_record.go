package usecases

import (
	"context"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type FindPetGrowthRecordInput struct {
	PetID  domain.PetID
	UserID domain.UserID
}

type FindPetGrowthRecordOutput struct {
	PetID            string
	Color            string
	CreatedAt        time.Time
	CurrentStageID   int
	Stages           []ActivePetEvolutionStageOutput
	TotalExperience  int64
	FeedCount        int
	ExperienceEvents []PetGrowthExperienceEventOutput
	Evolutions       []PetGrowthEvolutionOutput
}

type PetGrowthExperienceEventOutput struct {
	ID             string
	SourceType     string
	SourceID       *string
	Amount         int
	CappedAmount   int
	ExperienceDate time.Time
	CreatedAt      time.Time
}

type PetGrowthEvolutionOutput struct {
	ID              string
	StageID         int
	EvolutionRuleID *int
	PrimaryStatus   *string
	EvolvedAt       time.Time
	CreatedAt       time.Time
}

type FindPetGrowthRecord struct {
	petRepo             domain.PetRepository
	stageRepo           domain.EvolutionStageRepository
	experienceRepo      domain.PetExperienceRepository
	experienceEventRepo domain.PetExperienceEventRepository
	evolutionRepo       domain.PetEvolutionRepository
}

func NewFindPetGrowthRecord(
	petRepo domain.PetRepository,
	stageRepo domain.EvolutionStageRepository,
	experienceRepo domain.PetExperienceRepository,
	experienceEventRepo domain.PetExperienceEventRepository,
	evolutionRepo domain.PetEvolutionRepository,
) *FindPetGrowthRecord {
	return &FindPetGrowthRecord{
		petRepo:             petRepo,
		stageRepo:           stageRepo,
		experienceRepo:      experienceRepo,
		experienceEventRepo: experienceEventRepo,
		evolutionRepo:       evolutionRepo,
	}
}

func (u *FindPetGrowthRecord) Execute(ctx context.Context, input FindPetGrowthRecordInput) (FindPetGrowthRecordOutput, error) {
	if !domain.IsValidPetID(input.PetID) || !domain.IsValidUserID(input.UserID) {
		return FindPetGrowthRecordOutput{}, domain.ErrValidation
	}

	pet, err := u.petRepo.FindByID(ctx, input.PetID)
	if err != nil {
		return FindPetGrowthRecordOutput{}, err
	}
	if pet.UserID() != input.UserID {
		return FindPetGrowthRecordOutput{}, domain.ErrUnauthorized
	}

	experience, err := u.experienceRepo.FindByPetID(ctx, input.PetID)
	if err != nil {
		return FindPetGrowthRecordOutput{}, err
	}

	stages, err := u.stageRepo.FindAll(ctx)
	if err != nil {
		return FindPetGrowthRecordOutput{}, err
	}

	experienceEvents, err := u.experienceEventRepo.FindByPetID(ctx, input.PetID)
	if err != nil {
		return FindPetGrowthRecordOutput{}, err
	}

	evolutions, err := u.evolutionRepo.FindByPetID(ctx, input.PetID)
	if err != nil {
		return FindPetGrowthRecordOutput{}, err
	}

	return FindPetGrowthRecordOutput{
		PetID:            string(input.PetID),
		Color:            pet.Color(),
		CreatedAt:        pet.CreatedAt(),
		CurrentStageID:   pet.CurrentStageID(),
		Stages:           newActivePetEvolutionStageOutputs(stages, evolutions, pet.CurrentStageID()),
		TotalExperience:  experience.TotalExperience(),
		FeedCount:        experience.FeedCount(),
		ExperienceEvents: newPetGrowthExperienceEventOutputs(experienceEvents),
		Evolutions:       newPetGrowthEvolutionOutputs(evolutions),
	}, nil
}

func newPetGrowthExperienceEventOutputs(events []domain.PetExperienceEvent) []PetGrowthExperienceEventOutput {
	outputs := make([]PetGrowthExperienceEventOutput, 0, len(events))
	for _, event := range events {
		outputs = append(outputs, PetGrowthExperienceEventOutput{
			ID:             string(event.ID()),
			SourceType:     string(event.SourceType()),
			SourceID:       event.SourceID(),
			Amount:         event.Amount(),
			CappedAmount:   event.CappedAmount(),
			ExperienceDate: event.ExperienceDate(),
			CreatedAt:      event.CreatedAt(),
		})
	}
	return outputs
}

func newPetGrowthEvolutionOutputs(evolutions []domain.PetEvolution) []PetGrowthEvolutionOutput {
	outputs := make([]PetGrowthEvolutionOutput, 0, len(evolutions))
	for _, evolution := range evolutions {
		outputs = append(outputs, PetGrowthEvolutionOutput{
			ID:              string(evolution.ID()),
			StageID:         int(evolution.StageID()),
			EvolutionRuleID: evolutionRuleIDOutput(evolution.EvolutionRuleID()),
			PrimaryStatus:   evolution.PrimaryStatus(),
			EvolvedAt:       evolution.EvolvedAt(),
			CreatedAt:       evolution.CreatedAt(),
		})
	}
	return outputs
}

func evolutionRuleIDOutput(id *domain.EvolutionRuleID) *int {
	if id == nil {
		return nil
	}
	value := int(*id)
	return &value
}
