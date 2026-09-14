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
	output := p.Output(pet)
	return dto.PetResponse{
		ID:                   output.ID,
		Name:                 output.Name,
		Color:                output.Color,
		IsDeleted:            output.IsDeleted,
		UserID:               output.UserID,
		Energy:               output.Energy,
		Curiosity:            output.Curiosity,
		Sociality:            output.Sociality,
		Routine:              output.Routine,
		CurrentGroupMasterID: output.CurrentGroupMasterID,
		CurrentStageID:       output.CurrentStageID,
		CreatedAt:            output.CreatedAt,
		UpdatedAt:            output.UpdatedAt,
	}
}

func (p *PetPresenter) ListResponse(pets []domain.Pet) dto.PetListResponse {
	responses := make([]dto.PetResponse, 0, len(pets))
	for _, pet := range pets {
		responses = append(responses, p.Response(pet))
	}
	return dto.PetListResponse{Pets: responses}
}

func (p *PetPresenter) AllResponse(pets []domain.Pet) dto.AllPetListResponse {
	responses := make([]dto.AllPetResponse, 0, len(pets))
	for _, pet := range pets {
		responses = append(responses, dto.AllPetResponse{PetID: string(pet.ID()), Name: pet.Name()})
	}
	return dto.AllPetListResponse{Pets: responses}
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
