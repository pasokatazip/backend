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
	response := dto.LatestPetSouvenirResponse{}
	if output.Souvenir == nil {
		return response
	}

	response.Souvenir = &dto.LatestPetSouvenirItemResponse{
		ID:          output.Souvenir.ID,
		DisplayName: output.Souvenir.DisplayName,
		ImageURL:    output.Souvenir.ImageURL,
		FoundAt:     output.Souvenir.FoundAt,
		Reported:    output.Souvenir.Reported,
	}
	return response
}
