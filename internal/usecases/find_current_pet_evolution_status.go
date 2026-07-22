package usecases

import (
	"time"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/timeutil"
)

// FindCurrentPetEvolutionStatusInput identifies the authenticated pet whose
// evolution readiness is displayed. It never changes the pet's stage.
type FindCurrentPetEvolutionStatusInput struct {
	UserID domain.UserID
}

// FindCurrentPetEvolutionStatusOutput is intended for the evolution animation
// screen. can_evolve becomes true only for the branch that the server would
// choose when the evolution is committed.
type FindCurrentPetEvolutionStatusOutput struct {
	PetID        string                               `json:"pet_id"`
	CurrentStage CurrentPetEvolutionStageOutput       `json:"current_stage"`
	CanEvolve    bool                                 `json:"can_evolve"`
	NextStages   []CurrentPetEvolutionCandidateOutput `json:"next_stages"`
}

type CurrentPetEvolutionStageOutput struct {
	ID        int     `json:"id"`
	StageKey  string  `json:"stage_key"`
	StageNo   int     `json:"stage_no"`
	Name      string  `json:"name"`
	BranchKey *string `json:"branch_key,omitempty"`
	ImageURL  *string `json:"image_url,omitempty"`
}

type CurrentPetEvolutionCandidateOutput struct {
	RuleID         int                             `json:"rule_id"`
	ToStage        CurrentPetEvolutionStageOutput  `json:"to_stage"`
	SelectedForPet bool                            `json:"selected_for_pet"`
	Ready          bool                            `json:"ready"`
	Requirements   CurrentPetEvolutionRequirements `json:"requirements"`
}

type CurrentPetEvolutionRequirements struct {
	Experience             CurrentPetEvolutionProgress `json:"experience"`
	FeedCount              CurrentPetEvolutionProgress `json:"feed_count"`
	DaysSinceLastEvolution CurrentPetEvolutionProgress `json:"days_since_last_evolution"`
}

type CurrentPetEvolutionProgress struct {
	Current   int64 `json:"current"`
	Required  int64 `json:"required"`
	Remaining int64 `json:"remaining"`
	Met       bool  `json:"met"`
}

type FindCurrentPetEvolutionStatus struct {
	petRepo        domain.PetRepository
	experienceRepo domain.PetExperienceRepository
	stageRepo      domain.EvolutionStageRepository
	ruleRepo       domain.EvolutionRuleRepository
	evolutionRepo  domain.PetEvolutionRepository
}

func NewFindCurrentPetEvolutionStatus(
	petRepo domain.PetRepository,
	experienceRepo domain.PetExperienceRepository,
	stageRepo domain.EvolutionStageRepository,
	ruleRepo domain.EvolutionRuleRepository,
	evolutionRepo domain.PetEvolutionRepository,
) *FindCurrentPetEvolutionStatus {
	return &FindCurrentPetEvolutionStatus{
		petRepo:        petRepo,
		experienceRepo: experienceRepo,
		stageRepo:      stageRepo,
		ruleRepo:       ruleRepo,
		evolutionRepo:  evolutionRepo,
	}
}

// Execute calculates readiness only. The stage update remains in the command
// that awards experience, so a read request can never trigger an evolution.
func (u *FindCurrentPetEvolutionStatus) Execute(
	input FindCurrentPetEvolutionStatusInput,
) (FindCurrentPetEvolutionStatusOutput, error) {
	if !domain.IsValidUserID(input.UserID) {
		return FindCurrentPetEvolutionStatusOutput{}, domain.ErrValidation
	}

	pet, err := u.petRepo.FindActiveByUserID(input.UserID)
	if err != nil {
		return FindCurrentPetEvolutionStatusOutput{}, err
	}
	experience, err := u.experienceRepo.FindByPetID(pet.ID())
	if err != nil {
		return FindCurrentPetEvolutionStatusOutput{}, err
	}
	currentStage, err := u.stageRepo.FindByID(domain.EvolutionStageID(pet.CurrentStageID()))
	if err != nil {
		return FindCurrentPetEvolutionStatusOutput{}, err
	}
	rules, err := u.ruleRepo.FindByFromStageID(currentStage.ID())
	if err != nil {
		return FindCurrentPetEvolutionStatusOutput{}, err
	}
	evolutions, err := u.evolutionRepo.FindByPetID(pet.ID())
	if err != nil {
		return FindCurrentPetEvolutionStatusOutput{}, err
	}

	lastEvolvedAt := latestEvolutionTime(evolutions, pet.CreatedAt())
	now := timeutil.NowJST()
	candidates := make([]CurrentPetEvolutionCandidateOutput, 0, len(rules))
	canEvolve := false

	for _, rule := range rules {
		toStage, err := u.stageRepo.FindByID(rule.ToStageID())
		if err != nil {
			return FindCurrentPetEvolutionStatusOutput{}, err
		}

		selectedForPet := ruleCanApplyToPet(currentStage, toStage, pet)
		requirements := newCurrentPetEvolutionRequirements(rule, experience, lastEvolvedAt, now)
		ready := selectedForPet && requirementsMet(requirements)
		canEvolve = canEvolve || ready

		candidates = append(candidates, CurrentPetEvolutionCandidateOutput{
			RuleID:         int(rule.ID()),
			ToStage:        newCurrentPetEvolutionStageOutput(toStage),
			SelectedForPet: selectedForPet,
			Ready:          ready,
			Requirements:   requirements,
		})
	}

	return FindCurrentPetEvolutionStatusOutput{
		PetID:        string(pet.ID()),
		CurrentStage: newCurrentPetEvolutionStageOutput(currentStage),
		CanEvolve:    canEvolve,
		NextStages:   candidates,
	}, nil
}

func newCurrentPetEvolutionStageOutput(stage domain.EvolutionStage) CurrentPetEvolutionStageOutput {
	return CurrentPetEvolutionStageOutput{
		ID:        int(stage.ID()),
		StageKey:  stage.StageKey(),
		StageNo:   stage.StageNo(),
		Name:      stage.Name(),
		BranchKey: stage.BranchKey(),
		ImageURL:  stage.ImageURL(),
	}
}

// This mirrors the branch choice used when a feed actually commits an
// evolution. Keeping it here makes the preview's target match the animation.
func ruleCanApplyToPet(currentStage, toStage domain.EvolutionStage, pet domain.Pet) bool {
	if currentStage.ID() != 0 || toStage.BranchKey() == nil {
		return true
	}
	return *toStage.BranchKey() == evolutionBranchForPet(pet)
}

func evolutionBranchForPet(pet domain.Pet) string {
	switch {
	case pet.Curiosity() >= pet.Sociality() && pet.Curiosity() >= pet.Routine():
		return "shokushu"
	case pet.Sociality() >= pet.Routine():
		return "yonshoku"
	default:
		return "nishoku"
	}
}

func latestEvolutionTime(evolutions []domain.PetEvolution, fallback time.Time) time.Time {
	latest := fallback
	for _, evolution := range evolutions {
		if evolution.EvolvedAt().After(latest) {
			latest = evolution.EvolvedAt()
		}
	}
	return latest
}

func newCurrentPetEvolutionRequirements(
	rule domain.EvolutionRule,
	experience domain.PetExperience,
	lastEvolvedAt time.Time,
	now time.Time,
) CurrentPetEvolutionRequirements {
	days := int64(now.Sub(lastEvolvedAt).Hours() / 24)
	if days < 0 {
		days = 0
	}
	return CurrentPetEvolutionRequirements{
		Experience:             newCurrentPetEvolutionProgress(experience.TotalExperience(), rule.RequiredExperience()),
		FeedCount:              newCurrentPetEvolutionProgress(int64(experience.FeedCount()), int64(rule.RequiredFeedCount())),
		DaysSinceLastEvolution: newCurrentPetEvolutionProgress(days, int64(rule.RequiredDaysSinceLastEvolution())),
	}
}

func newCurrentPetEvolutionProgress(current, required int64) CurrentPetEvolutionProgress {
	remaining := required - current
	if remaining < 0 {
		remaining = 0
	}
	return CurrentPetEvolutionProgress{
		Current:   current,
		Required:  required,
		Remaining: remaining,
		Met:       current >= required,
	}
}

func requirementsMet(requirements CurrentPetEvolutionRequirements) bool {
	return requirements.Experience.Met &&
		requirements.FeedCount.Met &&
		requirements.DaysSinceLastEvolution.Met
}
