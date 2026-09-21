package usecases

import (
	"context"
	"testing"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type petGrowthRecordPetRepoStub struct {
	domain.PetRepository

	pet domain.Pet
}

func (s *petGrowthRecordPetRepoStub) Create(_ context.Context, pet domain.Pet) (domain.Pet, error) {
	return pet, nil
}

func (s *petGrowthRecordPetRepoStub) FindByID(_ context.Context, _ domain.PetID) (domain.Pet, error) {
	return s.pet, nil
}

func (s *petGrowthRecordPetRepoStub) FindActiveByUserID(_ context.Context, _ domain.UserID) (domain.Pet, error) {
	return domain.Pet{}, domain.ErrNotFound
}

func (s *petGrowthRecordPetRepoStub) FindAllByUserID(_ context.Context, _ domain.UserID) ([]domain.Pet, error) {
	return nil, nil
}

func (s *petGrowthRecordPetRepoStub) FindDeletedByUserID(_ context.Context, _ domain.UserID) ([]domain.Pet, error) {
	return nil, nil
}

func (s *petGrowthRecordPetRepoStub) UpdateProfile(_ context.Context,
	_ domain.PetID,
	_ domain.UserID,
	_ string,
	_ string,
	_ time.Time,
) (domain.Pet, error) {
	return domain.Pet{}, domain.ErrNotFound
}

type petGrowthRecordStageRepoStub struct{}

func (s *petGrowthRecordStageRepoStub) FindByID(_ context.Context, _ domain.EvolutionStageID) (domain.EvolutionStage, error) {
	return domain.EvolutionStage{}, domain.ErrNotFound
}

func (s *petGrowthRecordStageRepoStub) FindByStageNo(_ context.Context, _ int) (domain.EvolutionStage, error) {
	return domain.EvolutionStage{}, domain.ErrNotFound
}

func (s *petGrowthRecordStageRepoStub) FindAll(_ context.Context) ([]domain.EvolutionStage, error) {
	return nil, nil
}

type petGrowthRecordExperienceRepoStub struct {
	domain.PetExperienceRepository

	experience domain.PetExperience
}

func (s *petGrowthRecordExperienceRepoStub) Create(_ context.Context, experience domain.PetExperience) (domain.PetExperience, error) {
	return experience, nil
}

func (s *petGrowthRecordExperienceRepoStub) FindByPetID(_ context.Context, _ domain.PetID) (domain.PetExperience, error) {
	return s.experience, nil
}

func (s *petGrowthRecordExperienceRepoStub) Update(_ context.Context, experience domain.PetExperience) (domain.PetExperience, error) {
	return experience, nil
}

type petGrowthRecordExperienceEventRepoStub struct {
	domain.PetExperienceEventRepository
}

func (s *petGrowthRecordExperienceEventRepoStub) Create(_ context.Context, event domain.PetExperienceEvent) (domain.PetExperienceEvent, error) {
	return event, nil
}

func (s *petGrowthRecordExperienceEventRepoStub) FindByPetID(_ context.Context, _ domain.PetID) ([]domain.PetExperienceEvent, error) {
	return nil, nil
}

func (s *petGrowthRecordExperienceEventRepoStub) FindByPetIDAndDate(_ context.Context,
	_ domain.PetID,
	_ time.Time,
) ([]domain.PetExperienceEvent, error) {
	return nil, nil
}

type petGrowthRecordEvolutionRepoStub struct{ domain.PetEvolutionRepository }

func (s *petGrowthRecordEvolutionRepoStub) Create(_ context.Context, evolution domain.PetEvolution) (domain.PetEvolution, error) {
	return evolution, nil
}

func (s *petGrowthRecordEvolutionRepoStub) FindByPetID(_ context.Context, _ domain.PetID) ([]domain.PetEvolution, error) {
	return nil, nil
}

func (s *petGrowthRecordEvolutionRepoStub) FindLatestByPetID(_ context.Context, _ domain.PetID) (domain.PetEvolution, error) {
	return domain.PetEvolution{}, domain.ErrNotFound
}

func TestFindPetGrowthRecordIncludesPetMetadata(t *testing.T) {
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
	).Execute(context.Background(), FindPetGrowthRecordInput{PetID: petID, UserID: userID})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if output.Color != "#A1B2C3" {
		t.Fatalf("Color = %q, want %q", output.Color, "#A1B2C3")
	}
	if !output.CreatedAt.Equal(now) {
		t.Fatalf("CreatedAt = %s, want %s", output.CreatedAt, now)
	}

}
