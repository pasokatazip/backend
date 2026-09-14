package dto

import "time"

type FincodeCheckoutResponse struct {
	CheckoutURL string    `json:"checkout_url"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type FincodeSubscriptionStatusResponse struct {
	Active         bool    `json:"active"`
	CustomerID     *string `json:"fincode_customer_id,omitempty"`
	SubscriptionID *string `json:"fincode_subscription_id,omitempty"`
}

type FincodePurchaseConfirmResponse struct {
	Subsc bool `json:"subsc"`
}
