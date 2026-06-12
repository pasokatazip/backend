package usecases

import (
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type CreatePetInput struct {
	Name                 string
	UserID               string
	CurrentGroupMasterID *string
	CurrentStageID       string
}

type PetOutput struct {
	ID                   string
	Name                 string
	IsDeleted            bool
	UserID               string
	Energy               int
	Curiosity            int
	Sociality            int
	Routine              int
	CurrentGroupMasterID *string
	CurrentStageID       string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type CreatePet struct {
	repo domain.PetRepository
}

func NewCreatePet(repo domain.PetRepository) *CreatePet {
	return &CreatePet{repo: repo}
}

func (u *CreatePet) Execute(input CreatePetInput) (domain.Pet, error) {
	if input.Name == "" || input.UserID == "" || input.CurrentStageID == "" {
		return domain.Pet{}, domain.ErrValidation
	}

	if !domain.IsValidUserID(input.UserID) {
		return domain.Pet{}, domain.ErrValidation
	}

	if !domain.IsValidUserID(input.CurrentStageID) {
		return domain.Pet{}, domain.ErrValidation
	}
	
	if input.CurrentGroupMasterID != nil && !domain.IsValidUserID(*input.CurrentGroupMasterID) {
		return domain.Pet{}, domain.ErrValidation
	}

	now := time.Now().UTC()

	pet := domain.NewPet(
		domain.NewPetID(),
		input.Name,
		false,
		domain.UserID(input.UserID),
		0,
		0,
		0,
		0,
		input.CurrentGroupMasterID,
		input.CurrentStageID,
		now,
		now,
	)

	return u.repo.Create(pet)
}