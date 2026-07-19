package fincode

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pasokatazip/backend/internal/domain"
)

const fincodeDateLayout = "2006/01/02"

type createSubscriptionRequest struct {
	PayType    string `json:"pay_type"`
	PlanID     string `json:"plan_id"`
	CustomerID string `json:"customer_id"`
	CardID     string `json:"card_id,omitempty"`
	StartDate  string `json:"start_date"`
}

type createSubscriptionResponse struct {
	ID         string `json:"id"`
	PlanID     string `json:"plan_id"`
	CustomerID string `json:"customer_id"`
	Status     string `json:"status"`
}

func (c *Client) CreateSubscription(
	ctx context.Context,
	input domain.FincodeSubscriptionInput,
) (domain.FincodeSubscription, error) {
	if input.PlanID == "" || input.CustomerID == "" || input.StartDate.IsZero() {
		return domain.FincodeSubscription{}, domain.ErrValidation
	}

	var response createSubscriptionResponse
	err := c.doJSON(
		ctx,
		http.MethodPost,
		"/v1/subscriptions",
		input.IdempotencyKey,
		createSubscriptionRequest{
			PayType:    "Card",
			PlanID:     input.PlanID,
			CustomerID: input.CustomerID,
			CardID:     input.CardID,
			StartDate:  input.StartDate.Format(fincodeDateLayout),
		},
		&response,
	)
	if err != nil {
		return domain.FincodeSubscription{}, fmt.Errorf("create fincode subscription: %w", err)
	}
	if response.ID == "" {
		return domain.FincodeSubscription{}, fmt.Errorf("%w: create fincode subscription: response has no subscription id", domain.ErrExternalService)
	}

	return domain.FincodeSubscription{
		ID:         response.ID,
		CustomerID: response.CustomerID,
		PlanID:     response.PlanID,
		Status:     response.Status,
	}, nil
}

func (c *Client) CancelSubscription(ctx context.Context, subscriptionID string) error {
	if subscriptionID == "" {
		return domain.ErrValidation
	}

	path := "/v1/subscriptions/" + url.PathEscape(subscriptionID) + "?pay_type=Card"
	if err := c.doJSON(ctx, http.MethodDelete, path, "", nil, nil); err != nil {
		return fmt.Errorf("cancel fincode subscription: %w", err)
	}
	return nil
}
