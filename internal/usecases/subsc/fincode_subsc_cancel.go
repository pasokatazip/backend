package subsc

import (
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

func (u *SubscCancel) Execute(input SubscCancelInput) error {
	var (
		user domain.User
		err  error
	)

	switch {
	case input.SubscriptionID != "":
		user, err = u.repo.FindByFincodeSubscriptionID(input.SubscriptionID)
		if errors.Is(err, domain.ErrNotFound) && input.CustomerID != "" {
			user, err = u.repo.FindByFincodeCustomerID(input.CustomerID)
		}
	case input.CustomerID != "":
		user, err = u.repo.FindByFincodeCustomerID(input.CustomerID)
	default:
		return domain.ErrValidation
	}
	if err != nil {
		return err
	}

	return u.repo.UpdateSubscriptionStatus(user.ID(), false)
}
