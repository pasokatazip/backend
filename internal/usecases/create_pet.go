package usecases

import (
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type CreatePetInput struct {
	Name                 string
	UserID               domain.UserID
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
	CurrentStageID       int
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
	if input.Name == "" || !domain.IsValidUserID(input.UserID) {
		return domain.Pet{}, domain.ErrValidation
	}

	now := time.Now().UTC()

	pet := domain.NewPet(
		domain.NewPetID(),
		input.Name,
		false,
		input.UserID,
		0,	// Energy
		0,	// Curiosity
		0,	// Sociality
		0,	// Routine
		nil,	// CurrentGroupMasterID
		0,	// CurrentStageID
		now,
		now,
	)

	return u.repo.Create(pet)
}