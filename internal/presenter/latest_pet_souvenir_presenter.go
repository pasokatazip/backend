package presenter

import (
	"github.com/pasokatazip/backend/internal/controllers/dto"
	"github.com/pasokatazip/backend/internal/usecases"
)

type LatestPetSouvenirPresenter struct{}

func NewLatestPetSouvenirPresenter() *LatestPetSouvenirPresenter {
	return &LatestPetSouvenirPresenter{}
}

func (p *LatestPetSouvenirPresenter) Output(output usecases.FindLatestPetSouvenirOutput) dto.LatestPetSouvenirResponse {
	return dto.NewLatestPetSouvenirResponse(output)
}
