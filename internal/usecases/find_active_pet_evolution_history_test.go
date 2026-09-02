package usecases

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type activeEvolutionHistoryPetRepo struct {
	pet domain.Pet
}

func (r *activeEvolutionHistoryPetRepo) Create(pet domain.Pet) (domain.Pet, error) {
	return pet, nil
}

func (r *activeEvolutionHistoryPetRepo) FindByID(_ domain.PetID) (domain.Pet, error) {
	return r.pet, nil
}

func (r *activeEvolutionHistoryPetRepo) FindActiveByUserID(
	_ domain.UserID,
) (domain.Pet, error) {
	return r.pet, nil
}

func (r *activeEvolutionHistoryPetRepo) FindAllByUserID(
	_ domain.UserID,
) ([]domain.Pet, error) {
	return nil, nil
}

func (r *activeEvolutionHistoryPetRepo) FindDeletedByUserID(
	_ domain.UserID,
) ([]domain.Pet, error) {
	return nil, nil
}

func (r *activeEvolutionHistoryPetRepo) UpdateProfile(
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

func (r *activeEvolutionHistoryStageRepo) FindByID(
	_ domain.EvolutionStageID,
) (domain.EvolutionStage, error) {
	return domain.EvolutionStage{}, domain.ErrNotFound
}

func (r *activeEvolutionHistoryStageRepo) FindByStageNo(
	_ int,
) (domain.EvolutionStage, error) {
	return domain.EvolutionStage{}, domain.ErrNotFound
}

func (r *activeEvolutionHistoryStageRepo) FindAll() ([]domain.EvolutionStage, error) {
	return r.stages, nil
}

type activeEvolutionHistoryEvolutionRepo struct{}

func (r *activeEvolutionHistoryEvolutionRepo) Create(
	evolution domain.PetEvolution,
) (domain.PetEvolution, error) {
	return evolution, nil
}

func (r *activeEvolutionHistoryEvolutionRepo) FindByPetID(
	_ domain.PetID,
) ([]domain.PetEvolution, error) {
	return nil, nil
}

func (r *activeEvolutionHistoryEvolutionRepo) FindLatestByPetID(
	_ domain.PetID,
) (domain.PetEvolution, error) {
	return domain.PetEvolution{}, domain.ErrNotFound
}

func TestFindActivePetEvolutionHistoryReturnsStageKey(t *testing.T) {
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

	output, err := usecase.Execute(FindActivePetEvolutionHistoryInput{UserID: userID})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if output.StageKey != "amae_energy" {
		t.Fatalf("StageKey = %q, want %q", output.StageKey, "amae_energy")
	}

	encoded, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	var response map[string]any
	if err := json.Unmarshal(encoded, &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if response["stage_key"] != "amae_energy" {
		t.Fatalf("JSON stage_key = %v, want %q", response["stage_key"], "amae_energy")
	}
	if _, exists := response["current_stage_id"]; exists {
		t.Fatal("JSON must not include current_stage_id")
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

	_, err := usecase.Execute(FindActivePetEvolutionHistoryInput{UserID: userID})
	if !errors.Is(err, domain.ErrInternal) {
		t.Fatalf("error = %v, want ErrInternal", err)
	}
}
