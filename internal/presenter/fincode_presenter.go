package presenter

import (
	"github.com/pasokatazip/backend/internal/controllers/dto"
	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/usecases/onetime"
	"github.com/pasokatazip/backend/internal/usecases/subsc"
)

type FincodePresenter struct{}

func NewFincodePresenter() *FincodePresenter {
	return &FincodePresenter{}
}

func (p *FincodePresenter) Checkout(session domain.FincodeCardSession) dto.FincodeCheckoutResponse {
	return dto.FincodeCheckoutResponse{
		CheckoutURL: session.LinkURL,
		ExpiresAt:   session.ExpiresAt,
	}
}

func (p *FincodePresenter) SubscriptionStatus(status subsc.FincodeSubscriptionStatus) dto.FincodeSubscriptionStatusResponse {
	return dto.FincodeSubscriptionStatusResponse{
		Active:         status.Active,
		CustomerID:     status.CustomerID,
		SubscriptionID: status.SubscriptionID,
	}
}

func (p *FincodePresenter) PurchaseConfirm(status onetime.FincodePurchaseStatus) dto.FincodePurchaseConfirmResponse {
	return dto.FincodePurchaseConfirmResponse{Subsc: status.Purchased}
}
