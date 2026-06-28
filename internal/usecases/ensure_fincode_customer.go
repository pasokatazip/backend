package usecases

import (
	"context"
	"strings"

	"github.com/pasokatazip/backend/internal/domain"
)

type EnsureFincodeCustomer struct {
	repo    domain.UserRepository
	gateway domain.FincodeGateway
}

func NewEnsureFincodeCustomer(
	repo domain.UserRepository,
	gateway domain.FincodeGateway,
) *EnsureFincodeCustomer {
	return &EnsureFincodeCustomer{
		repo:    repo,
		gateway: gateway,
	}
}

func (u *EnsureFincodeCustomer) Execute(
	ctx context.Context,
	userID domain.UserID,
) (domain.FincodeCustomer, error) {
	if !domain.IsValidUserID(userID) || u.gateway == nil {
		return domain.FincodeCustomer{}, domain.ErrValidation
	}

	user, err := u.repo.FindByID(userID)
	if err != nil {
		return domain.FincodeCustomer{}, err
	}

	if customerID := user.FincodeCustomerID(); customerID != nil && strings.TrimSpace(*customerID) != "" {
		return domain.FincodeCustomer{
			ID:    *customerID,
			Email: user.Email(),
		}, nil
	}

	// UserID is a UUID v4, so it can also be used as fincode's idempotent_key.
	// Supplying it as the customer ID makes webhook-to-user association explicit.
	requestID := string(user.ID())
	customer, err := u.gateway.CreateCustomer(ctx, domain.FincodeCustomerInput{
		ID:             requestID,
		Email:          user.Email(),
		IdempotencyKey: requestID,
	})
	if err != nil {
		return domain.FincodeCustomer{}, err
	}

	if err := u.repo.UpdateFincodeCustomerID(user.ID(), customer.ID); err != nil {
		return domain.FincodeCustomer{}, err
	}

	return customer, nil
}
