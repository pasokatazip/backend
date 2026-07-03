package usecases

import "github.com/pasokatazip/backend/internal/domain"

type FindAllPetsInput struct {
	UserID domain.UserID
}

type FindAllPets struct {
	repo domain.PetRepository
}

func NewFindAllPets(repo domain.PetRepository) *FindAllPets {
	return &FindAllPets{repo: repo}
}

func (u *FindAllPets) Execute(input FindAllPetsInput) ([]domain.Pet, error) {
	if !domain.IsValidUserID(input.UserID) {
		return nil, domain.ErrValidation
	}

	return u.repo.FindAllByUserID(input.UserID)
}
