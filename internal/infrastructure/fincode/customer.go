package fincode

import (
	"context"
	"fmt"

	"github.com/pasokatazip/backend/internal/domain"
)

type createCustomerRequest struct {
	ID    string `json:"id,omitempty"`
	Email string `json:"email,omitempty"`
}

type createCustomerResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func (c *Client) CreateCustomer(
	ctx context.Context,
	input domain.FincodeCustomerInput,
) (domain.FincodeCustomer, error) {
	if input.Email == "" {
		return domain.FincodeCustomer{}, domain.ErrValidation
	}

	var response createCustomerResponse
	err := c.doJSON(
		ctx,
		"POST",
		"/v1/customers",
		input.IdempotencyKey,
		createCustomerRequest{
			ID:    input.ID,
			Email: input.Email,
		},
		&response,
	)
	if err != nil {
		return domain.FincodeCustomer{}, fmt.Errorf("create fincode customer: %w", err)
	}
	if response.ID == "" {
		return domain.FincodeCustomer{}, fmt.Errorf("create fincode customer: response has no customer id")
	}

	return domain.FincodeCustomer{
		ID:    response.ID,
		Email: response.Email,
	}, nil
}

var _ domain.FincodeGateway = (*Client)(nil)
