package usecases

import (
	"context"

	"github.com/pasokatazip/backend/internal/domain"
)

type FindAllPetsInput struct {
	UserID domain.UserID
}

type FindAllPets struct {
	repo domain.PetRepository
}

func NewFindAllPets(repo domain.PetRepository) *FindAllPets {
	return &FindAllPets{repo: repo}
}

func (u *FindAllPets) Execute(ctx context.Context, input FindAllPetsInput) ([]domain.Pet, error) {
	if !domain.IsValidUserID(input.UserID) {
		return nil, domain.ErrValidation
	}

	return u.repo.FindAllByUserID(ctx, input.UserID)
}
