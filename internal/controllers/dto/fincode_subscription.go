package dto

import (
	"time"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/usecases/subsc"
)

type FincodeCheckoutResponse struct {
	CheckoutURL string    `json:"checkout_url"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type FincodeSubscriptionStatusResponse struct {
	Active         bool    `json:"active"`
	CustomerID     *string `json:"fincode_customer_id,omitempty"`
	SubscriptionID *string `json:"fincode_subscription_id,omitempty"`
}

type FincodePurchaseStatusResponse struct {
	Purchased  bool    `json:"purchased"`
	CustomerID *string `json:"fincode_customer_id,omitempty"`
	PaymentID  *string `json:"fincode_payment_id,omitempty"`
}

func NewFincodeCheckoutResponse(session domain.FincodeCardSession) FincodeCheckoutResponse {
	return FincodeCheckoutResponse{
		CheckoutURL: session.LinkURL,
		ExpiresAt:   session.ExpiresAt,
	}
}

func NewFincodeSubscriptionStatusResponse(
	status subsc.FincodeSubscriptionStatus,
) FincodeSubscriptionStatusResponse {
	return FincodeSubscriptionStatusResponse{
		Active:         status.Active,
		CustomerID:     status.CustomerID,
		SubscriptionID: status.SubscriptionID,
	}
}
