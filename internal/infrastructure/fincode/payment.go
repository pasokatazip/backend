package fincode

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pasokatazip/backend/internal/domain"
)

type createPaymentRequest struct {
	ID      string `json:"id"`
	PayType string `json:"pay_type"`
	JobCode string `json:"job_code"`
	Amount  string `json:"amount"`
	Tax     string `json:"tax,omitempty"`
}

type executePaymentRequest struct {
	PayType    string `json:"pay_type"`
	JobCode    string `json:"job_code"`
	CustomerID string `json:"customer_id"`
	CardID     string `json:"card_id"`
	AccessID   string `json:"access_id"`
	Method     string `json:"method"`
}

type paymentResponse struct {
	ID       string `json:"id"`
	AccessID string `json:"access_id"`
	Status   string `json:"status"`
}

func (c *Client) GetPayment(ctx context.Context, paymentID string) (domain.FincodePayment, error) {
	if paymentID == "" {
		return domain.FincodePayment{}, domain.ErrValidation
	}

	var response paymentResponse
	err := c.doJSON(
		ctx,
		http.MethodGet,
		"/v1/payments/"+url.PathEscape(paymentID)+"?pay_type=Card",
		"",
		nil,
		&response,
	)
	if err != nil {
		return domain.FincodePayment{}, fmt.Errorf("get fincode payment: %w", err)
	}
	if response.ID == "" {
		return domain.FincodePayment{}, fmt.Errorf("%w: get fincode payment: response has no payment id", domain.ErrExternalService)
	}
	return domain.FincodePayment{
		ID: response.ID, AccessID: response.AccessID, Status: response.Status,
	}, nil
}

func (c *Client) CreatePayment(
	ctx context.Context,
	input domain.FincodePaymentInput,
) (domain.FincodePayment, error) {
	if input.ID == "" || input.Amount <= 0 {
		return domain.FincodePayment{}, domain.ErrValidation
	}

	var response paymentResponse
	err := c.doJSON(
		ctx,
		http.MethodPost,
		"/v1/payments",
		input.IdempotencyKey,
		createPaymentRequest{
			ID:      input.ID,
			PayType: "Card",
			JobCode: "CAPTURE",
			Amount:  fmt.Sprintf("%d", input.Amount),
		},
		&response,
	)
	if err != nil {
		if fincodeAPIErrorHasCode(err, "EC001025014") {
			return c.GetPayment(ctx, input.ID)
		}
		return domain.FincodePayment{}, fmt.Errorf("create fincode payment: %w", err)
	}
	if response.ID == "" {
		return domain.FincodePayment{}, fmt.Errorf("%w: create fincode payment: response has no payment id", domain.ErrExternalService)
	}
	return domain.FincodePayment{
		ID: response.ID, AccessID: response.AccessID, Status: response.Status,
	}, nil
}

func fincodeAPIErrorHasCode(err error, expected string) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	var response struct {
		Errors []struct {
			Code string `json:"error_code"`
		} `json:"errors"`
	}
	if json.Unmarshal([]byte(apiErr.Body), &response) != nil {
		return false
	}
	for _, item := range response.Errors {
		if item.Code == expected {
			return true
		}
	}
	return false
}

func (c *Client) ExecutePayment(
	ctx context.Context,
	input domain.FincodePaymentInput,
) (domain.FincodePayment, error) {
	if input.ID == "" || input.CustomerID == "" || input.CardID == "" || input.AccessID == "" {
		return domain.FincodePayment{}, domain.ErrValidation
	}

	var response paymentResponse
	err := c.doJSON(
		ctx,
		http.MethodPut,
		"/v1/payments/"+url.PathEscape(input.ID),
		input.IdempotencyKey,
		executePaymentRequest{
			PayType:    "Card",
			JobCode:    "CAPTURE",
			CustomerID: input.CustomerID,
			CardID:     input.CardID,
			AccessID:   input.AccessID,
			Method:     "1",
		},
		&response,
	)
	if err != nil {
		if fincodeAPIErrorHasCode(err, "EC002003002") {
			return domain.FincodePayment{}, fmt.Errorf(
				"execute fincode payment: customer no longer exists: %w",
				domain.ErrNotFound,
			)
		}
		return domain.FincodePayment{}, fmt.Errorf("execute fincode payment: %w", err)
	}
	if response.ID == "" {
		return domain.FincodePayment{}, fmt.Errorf("%w: execute fincode payment: response has no payment id", domain.ErrExternalService)
	}
	return domain.FincodePayment{
		ID: response.ID, AccessID: response.AccessID, Status: response.Status,
	}, nil
}
