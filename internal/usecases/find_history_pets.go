package usecases

import (
	"context"

	"github.com/pasokatazip/backend/internal/domain"
)

type FindHistoryPetsInput struct {
	UserID domain.UserID
}

type FindHistoryPets struct {
	repo domain.PetRepository
}

func NewFindHistoryPets(repo domain.PetRepository) *FindHistoryPets {
	return &FindHistoryPets{repo: repo}
}

func (u *FindHistoryPets) Execute(ctx context.Context, input FindHistoryPetsInput) ([]domain.Pet, error) {
	if !domain.IsValidUserID(input.UserID) {
		return nil, domain.ErrValidation
	}

	return u.repo.FindDeletedByUserID(ctx, input.UserID)
}
