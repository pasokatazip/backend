package presenter

import (
	"github.com/pasokatazip/backend/internal/controllers/dto"
	"github.com/pasokatazip/backend/internal/usecases"
)

type SouvenirPraiseFlagPresenter struct{}

func NewSouvenirPraiseFlagPresenter() *SouvenirPraiseFlagPresenter {
	return &SouvenirPraiseFlagPresenter{}
}

func (p *SouvenirPraiseFlagPresenter) Output(output usecases.SouvenirPraiseFlagOutput) dto.SouvenirPraiseFlagResponse {
	return dto.NewSouvenirPraiseFlagResponse(output)
}
