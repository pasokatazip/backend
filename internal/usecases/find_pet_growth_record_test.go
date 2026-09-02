package usecases

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type petGrowthRecordPetRepoStub struct {
	pet domain.Pet
}

func (s *petGrowthRecordPetRepoStub) Create(pet domain.Pet) (domain.Pet, error) {
	return pet, nil
}

func (s *petGrowthRecordPetRepoStub) FindByID(domain.PetID) (domain.Pet, error) {
	return s.pet, nil
}

func (s *petGrowthRecordPetRepoStub) FindActiveByUserID(domain.UserID) (domain.Pet, error) {
	return domain.Pet{}, domain.ErrNotFound
}

func (s *petGrowthRecordPetRepoStub) FindAllByUserID(domain.UserID) ([]domain.Pet, error) {
	return nil, nil
}

func (s *petGrowthRecordPetRepoStub) FindDeletedByUserID(domain.UserID) ([]domain.Pet, error) {
	return nil, nil
}

func (s *petGrowthRecordPetRepoStub) UpdateProfile(
	domain.PetID,
	domain.UserID,
	string,
	string,
	time.Time,
) (domain.Pet, error) {
	return domain.Pet{}, domain.ErrNotFound
}

type petGrowthRecordStageRepoStub struct{}

func (s *petGrowthRecordStageRepoStub) FindByID(domain.EvolutionStageID) (domain.EvolutionStage, error) {
	return domain.EvolutionStage{}, domain.ErrNotFound
}

func (s *petGrowthRecordStageRepoStub) FindByStageNo(int) (domain.EvolutionStage, error) {
	return domain.EvolutionStage{}, domain.ErrNotFound
}

func (s *petGrowthRecordStageRepoStub) FindAll() ([]domain.EvolutionStage, error) {
	return nil, nil
}

type petGrowthRecordExperienceRepoStub struct {
	experience domain.PetExperience
}

func (s *petGrowthRecordExperienceRepoStub) Create(experience domain.PetExperience) (domain.PetExperience, error) {
	return experience, nil
}

func (s *petGrowthRecordExperienceRepoStub) FindByPetID(domain.PetID) (domain.PetExperience, error) {
	return s.experience, nil
}

func (s *petGrowthRecordExperienceRepoStub) Update(experience domain.PetExperience) (domain.PetExperience, error) {
	return experience, nil
}

type petGrowthRecordExperienceEventRepoStub struct{}

func (s *petGrowthRecordExperienceEventRepoStub) Create(event domain.PetExperienceEvent) (domain.PetExperienceEvent, error) {
	return event, nil
}

func (s *petGrowthRecordExperienceEventRepoStub) FindByPetID(domain.PetID) ([]domain.PetExperienceEvent, error) {
	return nil, nil
}

func (s *petGrowthRecordExperienceEventRepoStub) FindByPetIDAndDate(
	domain.PetID,
	time.Time,
) ([]domain.PetExperienceEvent, error) {
	return nil, nil
}

type petGrowthRecordEvolutionRepoStub struct{}

func (s *petGrowthRecordEvolutionRepoStub) Create(evolution domain.PetEvolution) (domain.PetEvolution, error) {
	return evolution, nil
}

func (s *petGrowthRecordEvolutionRepoStub) FindByPetID(domain.PetID) ([]domain.PetEvolution, error) {
	return nil, nil
}

func (s *petGrowthRecordEvolutionRepoStub) FindLatestByPetID(domain.PetID) (domain.PetEvolution, error) {
	return domain.PetEvolution{}, domain.ErrNotFound
}

func TestFindPetGrowthRecordIncludesPetColor(t *testing.T) {
	userID := domain.UserID("4c0c926e-6a13-4bf4-8ae4-593c4047280f")
	petID := domain.PetID("b5d213dd-75f7-4bb2-b260-7efb4c04758a")
	now := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	pet := domain.NewPet(
		petID, "ぽち", "#A1B2C3", false, userID,
		50, 50, 50, 50, nil, 2, now, now,
	)
	experience := domain.NewPetExperience("experience-id", petID, 100, 5, now, now)

	output, err := NewFindPetGrowthRecord(
		&petGrowthRecordPetRepoStub{pet: pet},
		&petGrowthRecordStageRepoStub{},
		&petGrowthRecordExperienceRepoStub{experience: experience},
		&petGrowthRecordExperienceEventRepoStub{},
		&petGrowthRecordEvolutionRepoStub{},
	).Execute(FindPetGrowthRecordInput{PetID: petID, UserID: userID})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if output.Color != "#A1B2C3" {
		t.Fatalf("Color = %q, want %q", output.Color, "#A1B2C3")
	}

	encoded, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	var response map[string]any
	if err := json.Unmarshal(encoded, &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if response["color"] != "#A1B2C3" {
		t.Fatalf("JSON color = %v, want %q", response["color"], "#A1B2C3")
	}
}
