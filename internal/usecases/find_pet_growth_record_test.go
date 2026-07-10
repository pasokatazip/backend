package usecases

import (
	"errors"
	"testing"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type petGrowthExperienceRepository struct {
	petID      domain.PetID
	experience domain.PetExperience
}

func (r *petGrowthExperienceRepository) Create(petExperience domain.PetExperience) (domain.PetExperience, error) {
	return petExperience, nil
}

func (r *petGrowthExperienceRepository) FindByPetID(petID domain.PetID) (domain.PetExperience, error) {
	r.petID = petID
	return r.experience, nil
}

func (r *petGrowthExperienceRepository) Update(petExperience domain.PetExperience) (domain.PetExperience, error) {
	return petExperience, nil
}

type petGrowthExperienceEventRepository struct {
	petID  domain.PetID
	events []domain.PetExperienceEvent
}

func (r *petGrowthExperienceEventRepository) Create(event domain.PetExperienceEvent) (domain.PetExperienceEvent, error) {
	return event, nil
}

func (r *petGrowthExperienceEventRepository) FindByPetID(petID domain.PetID) ([]domain.PetExperienceEvent, error) {
	r.petID = petID
	return r.events, nil
}

func (r *petGrowthExperienceEventRepository) FindByPetIDAndDate(domain.PetID, time.Time) ([]domain.PetExperienceEvent, error) {
	return nil, nil
}

type petGrowthEvolutionRepository struct {
	petID      domain.PetID
	evolutions []domain.PetEvolution
}

func (r *petGrowthEvolutionRepository) Create(evolution domain.PetEvolution) (domain.PetEvolution, error) {
	return evolution, nil
}

func (r *petGrowthEvolutionRepository) FindByPetID(petID domain.PetID) ([]domain.PetEvolution, error) {
	r.petID = petID
	return r.evolutions, nil
}

func (r *petGrowthEvolutionRepository) FindLatestByPetID(domain.PetID) (domain.PetEvolution, error) {
	return domain.PetEvolution{}, nil
}

func TestFindPetGrowthRecord(t *testing.T) {
	petID := domain.PetID("d9428888-122b-11e1-b85c-61cd3cbb3210")
	now := time.Date(2026, 7, 10, 9, 0, 0, 0, time.UTC)
	sourceID := "post-id"
	ruleID := domain.EvolutionRuleID(1)
	primaryStatus := "curiosity"

	experienceRepo := &petGrowthExperienceRepository{
		experience: domain.NewPetExperience(
			domain.PetExperienceID("experience-id"),
			petID,
			120,
			3,
			now,
			now,
		),
	}
	eventRepo := &petGrowthExperienceEventRepository{
		events: []domain.PetExperienceEvent{
			domain.NewPetExperienceEvent(
				domain.PetExperienceEventID("event-id"),
				petID,
				domain.ExperienceSourceTypeFeed,
				&sourceID,
				20,
				0,
				now,
				now,
			),
		},
	}
	evolutionRepo := &petGrowthEvolutionRepository{
		evolutions: []domain.PetEvolution{
			domain.NewPetEvolution(
				domain.PetEvolutionID("evolution-id"),
				petID,
				domain.EvolutionStageID(2),
				&ruleID,
				&primaryStatus,
				now,
				now,
			),
		},
	}

	output, err := NewFindPetGrowthRecord(experienceRepo, eventRepo, evolutionRepo).Execute(
		FindPetGrowthRecordInput{PetID: petID},
	)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if experienceRepo.petID != petID || eventRepo.petID != petID || evolutionRepo.petID != petID {
		t.Fatalf("repository pet IDs = %q, %q, %q; want %q", experienceRepo.petID, eventRepo.petID, evolutionRepo.petID, petID)
	}
	if output.TotalExperience != 120 || output.FeedCount != 3 {
		t.Fatalf("output aggregate = %+v, want total 120/feed 3", output)
	}
	if len(output.ExperienceEvents) != 1 || output.ExperienceEvents[0].SourceType != "feed" {
		t.Fatalf("experience events = %+v, want one feed event", output.ExperienceEvents)
	}
	if len(output.Evolutions) != 1 || output.Evolutions[0].StageID != 2 {
		t.Fatalf("evolutions = %+v, want one stage 2 evolution", output.Evolutions)
	}
}

func TestFindPetGrowthRecordRejectsInvalidPetID(t *testing.T) {
	_, err := NewFindPetGrowthRecord(
		&petGrowthExperienceRepository{},
		&petGrowthExperienceEventRepository{},
		&petGrowthEvolutionRepository{},
	).Execute(FindPetGrowthRecordInput{PetID: "invalid"})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("error = %v, want ErrValidation", err)
	}
}
