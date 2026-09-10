package presenter

import (
	"github.com/pasokatazip/backend/internal/controllers/dto"
	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/usecases"
)

type PetPresenter struct{}

func NewPetPresenter() *PetPresenter {
	return &PetPresenter{}
}

func (p *PetPresenter) Response(pet domain.Pet) dto.PetResponse {
	return dto.NewPetResponse(p.Output(pet))
}

func (p *PetPresenter) ListResponse(pets []domain.Pet) dto.PetListResponse {
	outputs := make([]usecases.PetOutput, 0, len(pets))
	for _, pet := range pets {
		outputs = append(outputs, p.Output(pet))
	}
	return dto.NewPetListResponse(outputs)
}

func (p *PetPresenter) AllResponse(pets []domain.Pet) dto.AllPetListResponse {
	outputs := make([]usecases.PetOutput, 0, len(pets))
	for _, pet := range pets {
		outputs = append(outputs, p.Output(pet))
	}
	return dto.NewAllPetListResponse(outputs)
}

func (p *PetPresenter) Output(pet domain.Pet) usecases.PetOutput {
	return usecases.PetOutput{
		ID:                   string(pet.ID()),
		Name:                 pet.Name(),
		Color:                pet.Color(),
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
