package subsc

import (
	"context"
	"errors"

	"github.com/pasokatazip/backend/internal/domain"
)

type SubscCancelInput struct {
	CustomerID     string
	SubscriptionID string
}

type SubscCancel struct {
	repo domain.UserRepository
}

func NewSubscCancel(repo domain.UserRepository) *SubscCancel {
	return &SubscCancel{repo: repo}
}

func (u *SubscCancel) Execute(ctx context.Context, input SubscCancelInput) error {
	var (
		user domain.User
		err  error
	)

	switch {
	case input.SubscriptionID != "":
		user, err = u.repo.FindByFincodeSubscriptionID(ctx, input.SubscriptionID)
		if errors.Is(err, domain.ErrNotFound) && input.CustomerID != "" {
			user, err = u.repo.FindByFincodeCustomerID(ctx, input.CustomerID)
		}
	case input.CustomerID != "":
		user, err = u.repo.FindByFincodeCustomerID(ctx, input.CustomerID)
	default:
		return domain.ErrValidation
	}
	if err != nil {
		return err
	}

	return u.repo.UpdateSubscriptionStatus(ctx, user.ID(), false)
}
