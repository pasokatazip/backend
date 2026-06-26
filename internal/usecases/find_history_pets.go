package usecases

import "github.com/pasokatazip/backend/internal/domain"

type FindHistoryPetsInput struct {
	UserID domain.UserID
}

type FindHistoryPets struct {
	repo domain.PetRepository
}

func NewFindHistoryPets(repo domain.PetRepository) *FindHistoryPets {
	return &FindHistoryPets{repo: repo}
}

func (u *FindHistoryPets) Execute(input FindHistoryPetsInput) ([]domain.Pet, error) {
	if !domain.IsValidUserID(input.UserID) {
		return nil, domain.ErrValidation
	}

	return u.repo.FindDeletedByUserID(input.UserID)
}
