package presenter

import (
	"github.com/pasokatazip/backend/internal/controllers/dto"
	"github.com/pasokatazip/backend/internal/usecases"
)

type CurrentPetPresenter struct{}

func NewCurrentPetPresenter() *CurrentPetPresenter {
	return &CurrentPetPresenter{}
}

func (p *CurrentPetPresenter) Output(output usecases.FindMyActivePetOutput) dto.CurrentPetResponse {
	return dto.NewCurrentPetResponse(output)
}
