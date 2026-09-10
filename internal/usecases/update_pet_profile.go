package usecases

import (
	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/timeutil"
)

type UpdatePetProfileInput struct {
	PetID  domain.PetID
	UserID domain.UserID
	Name   string
	Color  string
}

type UpdatePetProfile struct {
	repo domain.PetRepository
}

func NewUpdatePetProfile(repo domain.PetRepository) *UpdatePetProfile {
	return &UpdatePetProfile{repo: repo}
}

func (u *UpdatePetProfile) Execute(input UpdatePetProfileInput) (domain.Pet, error) {
	name, nameOK := normalizeAndValidatePetName(input.Name)
	if !domain.IsValidPetID(input.PetID) ||
		!domain.IsValidUserID(input.UserID) ||
		!nameOK ||
		!domain.IsValidPetColor(input.Color) {
		return domain.Pet{}, domain.ErrValidation
	}

	return u.repo.UpdateProfile(input.PetID, input.UserID, name, input.Color, timeutil.NowJST())
}
