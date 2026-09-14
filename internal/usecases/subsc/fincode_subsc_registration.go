package subsc

import (
	"context"
	"strings"

	"github.com/pasokatazip/backend/internal/domain"
)

type SubscRegistrationInput struct {
	CustomerID     string
	SubscriptionID string
	Status         string
}

type SubscRegistration struct {
	repo domain.UserRepository
}

func NewSubscRegistration(repo domain.UserRepository) *SubscRegistration {
	return &SubscRegistration{repo: repo}
}

func (u *SubscRegistration) Execute(ctx context.Context, input SubscRegistrationInput) error {
	if input.CustomerID == "" || input.SubscriptionID == "" {
		return domain.ErrValidation
	}

	subsc, ok := subscriptionEnabled(input.Status)
	if !ok {
		return domain.ErrValidation
	}

	user, err := u.repo.FindByFincodeCustomerID(ctx, input.CustomerID)
	if err != nil {
		return err
	}

	return u.repo.UpdateFincodeSubscription(ctx, user.ID(), input.SubscriptionID, subsc)
}

func subscriptionEnabled(status string) (bool, bool) {
	switch strings.ToUpper(status) {
	case "ACTIVE", "RUNNING":
		return true, true
	case "CANCELED", "INCOMPLETE":
		return false, true
	default:
		return false, false
	}
}
