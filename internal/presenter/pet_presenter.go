package presenter

import (
	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/usecases"
)

type PetPresenter struct{}

func NewPetPresenter() *PetPresenter {
	return &PetPresenter{}
}

func (p *PetPresenter) Output(pet domain.Pet) usecases.PetOutput {
	return usecases.PetOutput{
		ID:                   string(pet.ID()),
		Name:                 pet.Name(),
		IsDeleted:            pet.IsDeleted(),
		UserID:               string(pet.UserID()),
		Energy:               pet.Energy(),
		Curiosity:            pet.Curiosity(),
		Sociality:            pet.Sociality(),
		Routine:              pet.Routine(),
		CurrentGroupMasterID: pet.CurrentGroupMasterID(),
		CurrentStageID:       pet.CurrentStageID(),
		CreatedAt:            pet.CreatedAt(),
		UpdatedAt:            pet.UpdatedAt(),
	}
}