package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type activeEvolutionHistoryPetRepo struct {
	domain.PetRepository

	pet domain.Pet
}

func (r *activeEvolutionHistoryPetRepo) Create(_ context.Context, pet domain.Pet) (domain.Pet, error) {
	return pet, nil
}

func (r *activeEvolutionHistoryPetRepo) FindByID(_ context.Context, _ domain.PetID) (domain.Pet, error) {
	return r.pet, nil
}

func (r *activeEvolutionHistoryPetRepo) FindActiveByUserID(_ context.Context,
	_ domain.UserID,
) (domain.Pet, error) {
	return r.pet, nil
}

func (r *activeEvolutionHistoryPetRepo) FindAllByUserID(_ context.Context,
	_ domain.UserID,
) ([]domain.Pet, error) {
	return nil, nil
}

func (r *activeEvolutionHistoryPetRepo) FindDeletedByUserID(_ context.Context,
	_ domain.UserID,
) ([]domain.Pet, error) {
	return nil, nil
}

func (r *activeEvolutionHistoryPetRepo) UpdateProfile(_ context.Context,
	_ domain.PetID,
	_ domain.UserID,
	_ string,
	_ string,
	_ time.Time,
) (domain.Pet, error) {
	return r.pet, nil
}

type activeEvolutionHistoryStageRepo struct {
	stages []domain.EvolutionStage
}

func (r *activeEvolutionHistoryStageRepo) FindByID(_ context.Context,
	_ domain.EvolutionStageID,
) (domain.EvolutionStage, error) {
	return domain.EvolutionStage{}, domain.ErrNotFound
}

func (r *activeEvolutionHistoryStageRepo) FindByStageNo(_ context.Context,
	_ int,
) (domain.EvolutionStage, error) {
	return domain.EvolutionStage{}, domain.ErrNotFound
}

func (r *activeEvolutionHistoryStageRepo) FindAll(_ context.Context) ([]domain.EvolutionStage, error) {
	return r.stages, nil
}

type activeEvolutionHistoryEvolutionRepo struct{ domain.PetEvolutionRepository }

func (r *activeEvolutionHistoryEvolutionRepo) Create(_ context.Context,
	evolution domain.PetEvolution,
) (domain.PetEvolution, error) {
	return evolution, nil
}

func (r *activeEvolutionHistoryEvolutionRepo) FindByPetID(_ context.Context,
	_ domain.PetID,
) ([]domain.PetEvolution, error) {
	return nil, nil
}

func (r *activeEvolutionHistoryEvolutionRepo) FindLatestByPetID(_ context.Context,
	_ domain.PetID,
) (domain.PetEvolution, error) {
	return domain.PetEvolution{}, domain.ErrNotFound
}

func TestFindActivePetEvolutionHistoryReturnsCurrentStageKey(t *testing.T) {
	userID := domain.UserID("4c0c926e-6a13-4bf4-8ae4-593c4047280f")
	petID := domain.PetID("b5d213dd-75f7-4bb2-b260-7efb4c04758a")
	now := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	pet := domain.NewPet(
		petID, "ぽち", "#A1B2C3", false, userID,
		50, 50, 50, 50, nil, 2, now, now,
	)
	stage := domain.NewEvolutionStage(
		2, "amae_energy", 2, "あまえんぼ", nil, nil, now, now,
	)
	usecase := NewFindActivePetEvolutionHistory(
		&activeEvolutionHistoryPetRepo{pet: pet},
		&activeEvolutionHistoryStageRepo{stages: []domain.EvolutionStage{stage}},
		&activeEvolutionHistoryEvolutionRepo{},
	)

	output, err := usecase.Execute(context.Background(), FindActivePetEvolutionHistoryInput{UserID: userID})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if output.CurrentStageKey != "amae_energy" {
		t.Fatalf("CurrentStageKey = %q, want %q", output.CurrentStageKey, "amae_energy")
	}

}

func TestFindActivePetEvolutionHistoryRejectsMissingCurrentStage(t *testing.T) {
	userID := domain.UserID("4c0c926e-6a13-4bf4-8ae4-593c4047280f")
	petID := domain.PetID("b5d213dd-75f7-4bb2-b260-7efb4c04758a")
	now := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	pet := domain.NewPet(
		petID, "ぽち", "#A1B2C3", false, userID,
		50, 50, 50, 50, nil, 99, now, now,
	)
	usecase := NewFindActivePetEvolutionHistory(
		&activeEvolutionHistoryPetRepo{pet: pet},
		&activeEvolutionHistoryStageRepo{},
		&activeEvolutionHistoryEvolutionRepo{},
	)

	_, err := usecase.Execute(context.Background(), FindActivePetEvolutionHistoryInput{UserID: userID})
	if !errors.Is(err, domain.ErrInternal) {
		t.Fatalf("error = %v, want ErrInternal", err)
	}
}
