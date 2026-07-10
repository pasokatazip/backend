package usecases

import "github.com/pasokatazip/backend/internal/domain"

type FindActivePetEvolutionHistoryInput struct {
	UserID domain.UserID
}

type FindActivePetEvolutionHistoryOutput struct {
	PetID      string                     `json:"pet_id"`
	Evolutions []PetGrowthEvolutionOutput `json:"evolutions"`
}

type FindActivePetEvolutionHistory struct {
	petRepo       domain.PetRepository
	evolutionRepo domain.PetEvolutionRepository
}

func NewFindActivePetEvolutionHistory(
	petRepo domain.PetRepository,
	evolutionRepo domain.PetEvolutionRepository,
) *FindActivePetEvolutionHistory {
	return &FindActivePetEvolutionHistory{
		petRepo:       petRepo,
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

	evolutions, err := u.evolutionRepo.FindByPetID(pet.ID())
	if err != nil {
		return FindActivePetEvolutionHistoryOutput{}, err
	}

	return FindActivePetEvolutionHistoryOutput{
		PetID:      string(pet.ID()),
		Evolutions: newPetGrowthEvolutionOutputs(evolutions),
	}, nil
}
