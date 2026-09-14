package usecases

import (
	"context"

	"github.com/pasokatazip/backend/internal/domain"
)

type FindLatestHistoricalPetSouvenirInput struct {
	UserID domain.UserID
	PetID  domain.PetID
}

type FindLatestHistoricalPetSouvenir struct {
	repo domain.PetSouvenirRepository
}

func NewFindLatestHistoricalPetSouvenir(
	repo domain.PetSouvenirRepository,
) *FindLatestHistoricalPetSouvenir {
	return &FindLatestHistoricalPetSouvenir{repo: repo}
}

func (u *FindLatestHistoricalPetSouvenir) Execute(ctx context.Context,
	input FindLatestHistoricalPetSouvenirInput,
) (FindLatestPetSouvenirOutput, error) {
	if !domain.IsValidUserID(input.UserID) || !domain.IsValidPetID(input.PetID) {
		return FindLatestPetSouvenirOutput{}, domain.ErrValidation
	}

	souvenir, err := u.repo.FindLatestByHistoricalPetID(ctx, input.UserID, input.PetID)
	if err != nil {
		return FindLatestPetSouvenirOutput{}, err
	}
	if souvenir == nil {
		return FindLatestPetSouvenirOutput{}, nil
	}

	return FindLatestPetSouvenirOutput{
		Souvenir: newLatestPetSouvenirOutput(*souvenir),
	}, nil
}
