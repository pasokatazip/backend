package usecases

import (
	"errors"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/timeutil"
)

type CreatePetInput struct {
	Name   string
	Color  string
	UserID domain.UserID
}

type PetOutput struct {
	ID                   string
	Name                 string
	Color                string
	IsDeleted            bool
	UserID               string
	Energy               float64
	Curiosity            float64
	Sociality            float64
	Routine              float64
	CurrentGroupMasterID *int
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

	color := input.Color
	if color == "" {
		color = domain.DefaultPetColor
	}
	if !domain.IsValidPetColor(color) {
		return domain.Pet{}, domain.ErrValidation
	}

	// user_active_pets はユーザーごと1件だけにする。
	// Setupへの直接アクセスや二重送信でも、2匹目を作成しない。
	if _, err := u.repo.FindActiveByUserID(input.UserID); err == nil {
		return domain.Pet{}, domain.ErrAlreadyExists
	} else if !errors.Is(err, domain.ErrNotFound) {
		return domain.Pet{}, err
	}

	now := timeutil.NowJST()

	pet := domain.NewPet(
		domain.NewPetID(),
		input.Name,
		color,
		false,
		input.UserID,
		50,  // Energy
		50,  // Curiosity
		50,  // Sociality
		50,  // Routine
		nil, // CurrentGroupMasterID
		0,   // CurrentStageID
		now,
		now,
	)

	return u.repo.Create(pet)
}
