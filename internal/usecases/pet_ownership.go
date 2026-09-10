package usecases

import "github.com/pasokatazip/backend/internal/domain"

func findOwnedPet(
	petRepo domain.PetRepository,
	userID domain.UserID,
	petID domain.PetID,
) (domain.Pet, error) {
	if !domain.IsValidUserID(userID) || !domain.IsValidPetID(petID) {
		return domain.Pet{}, domain.ErrValidation
	}

	pet, err := petRepo.FindByID(petID)
	if err != nil {
		return domain.Pet{}, err
	}
	if pet.UserID() != userID {
		return domain.Pet{}, domain.ErrUnauthorized
	}

	return pet, nil
}
