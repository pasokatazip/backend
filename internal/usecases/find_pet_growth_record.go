package usecases

import (
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type FindPetGrowthRecordInput struct {
	PetID  domain.PetID
	UserID domain.UserID
}

type FindPetGrowthRecordOutput struct {
	PetID            string                           `json:"pet_id"`
	CurrentStageID   int                              `json:"current_stage_id"`
	Stages           []ActivePetEvolutionStageOutput  `json:"stages"`
	TotalExperience  int64                            `json:"total_experience"`
	FeedCount        int                              `json:"feed_count"`
	ExperienceEvents []PetGrowthExperienceEventOutput `json:"experience_events"`
	Evolutions       []PetGrowthEvolutionOutput       `json:"evolutions"`
}

type PetGrowthExperienceEventOutput struct {
	ID             string    `json:"id"`
	SourceType     string    `json:"source_type"`
	SourceID       *string   `json:"source_id,omitempty"`
	Amount         int       `json:"amount"`
	CappedAmount   int       `json:"capped_amount"`
	ExperienceDate time.Time `json:"experience_date"`
	CreatedAt      time.Time `json:"created_at"`
}

type PetGrowthEvolutionOutput struct {
	ID              string    `json:"id"`
	StageID         int       `json:"stage_id"`
	EvolutionRuleID *int      `json:"evolution_rule_id,omitempty"`
	PrimaryStatus   *string   `json:"primary_status,omitempty"`
	EvolvedAt       time.Time `json:"evolved_at"`
	CreatedAt       time.Time `json:"created_at"`
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

func (u *FindPetGrowthRecord) Execute(input FindPetGrowthRecordInput) (FindPetGrowthRecordOutput, error) {
	if !domain.IsValidPetID(input.PetID) || !domain.IsValidUserID(input.UserID) {
		return FindPetGrowthRecordOutput{}, domain.ErrValidation
	}

	pet, err := u.petRepo.FindByID(input.PetID)
	if err != nil {
		return FindPetGrowthRecordOutput{}, err
	}
	if pet.UserID() != input.UserID {
		return FindPetGrowthRecordOutput{}, domain.ErrUnauthorized
	}

	experience, err := u.experienceRepo.FindByPetID(input.PetID)
	if err != nil {
		return FindPetGrowthRecordOutput{}, err
	}

	stages, err := u.stageRepo.FindAll()
	if err != nil {
		return FindPetGrowthRecordOutput{}, err
	}

	experienceEvents, err := u.experienceEventRepo.FindByPetID(input.PetID)
	if err != nil {
		return FindPetGrowthRecordOutput{}, err
	}

	evolutions, err := u.evolutionRepo.FindByPetID(input.PetID)
	if err != nil {
		return FindPetGrowthRecordOutput{}, err
	}

	return FindPetGrowthRecordOutput{
		PetID:            string(input.PetID),
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
