package usecases

import (
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type FindLatestPetSouvenirInput struct {
	UserID domain.UserID
}

type LatestPetSouvenirOutput struct {
	ID          string
	DisplayName string
	ImageURL    string
	FoundAt     time.Time
	Reported    bool
}

type FindLatestPetSouvenirOutput struct {
	Souvenir *LatestPetSouvenirOutput
}

type FindLatestPetSouvenir struct {
	repo domain.PetSouvenirRepository
}

func NewFindLatestPetSouvenir(
	repo domain.PetSouvenirRepository,
) *FindLatestPetSouvenir {
	return &FindLatestPetSouvenir{repo: repo}
}

func (u *FindLatestPetSouvenir) Execute(
	input FindLatestPetSouvenirInput,
) (FindLatestPetSouvenirOutput, error) {
	if !domain.IsValidUserID(input.UserID) {
		return FindLatestPetSouvenirOutput{}, domain.ErrValidation
	}

	souvenir, err := u.repo.FindLatestByActivePetUserID(input.UserID)
	if err != nil {
		return FindLatestPetSouvenirOutput{}, err
	}
	if souvenir == nil {
		return FindLatestPetSouvenirOutput{}, nil
	}

	return FindLatestPetSouvenirOutput{
		Souvenir: &LatestPetSouvenirOutput{
			ID:          souvenir.ID(),
			DisplayName: souvenir.DisplayName(),
			ImageURL:    souvenir.ImageURL(),
			FoundAt:     souvenir.FoundAt(),
			Reported:    souvenir.Reported(),
		},
	}, nil
}
