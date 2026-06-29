package usecases

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/timeutil"
)

type CardRegistrationInput struct {
	CustomerID string
	CardID     string
}

type CardRegistration struct {
	repo    domain.UserRepository
	gateway domain.FincodeGateway
	planID  string
	now     func() time.Time
}

func NewCardRegistration(
	repo domain.UserRepository,
	gateway domain.FincodeGateway,
	planID string,
) *CardRegistration {
	return &CardRegistration{
		repo:    repo,
		gateway: gateway,
		planID:  planID,
		now:     timeutil.NowJST,
	}
}

func (u *CardRegistration) Execute(ctx context.Context, input CardRegistrationInput) error {
	if input.CustomerID == "" || input.CardID == "" || strings.TrimSpace(u.planID) == "" || u.gateway == nil {
		return domain.ErrValidation
	}

	user, err := u.repo.FindByFincodeCustomerID(input.CustomerID)
	if errors.Is(err, domain.ErrNotFound) && domain.IsValidUserID(domain.UserID(input.CustomerID)) {
		if err := u.repo.UpdateFincodeCustomerID(domain.UserID(input.CustomerID), input.CustomerID); err != nil {
			return err
		}
		user, err = u.repo.FindByID(domain.UserID(input.CustomerID))
	}
	if err != nil {
		return err
	}
	if user.Subsc() && user.FincodeSubscriptionID() != nil && *user.FincodeSubscriptionID() != "" {
		return nil
	}

	subscription, err := u.gateway.CreateSubscription(ctx, domain.FincodeSubscriptionInput{
		PlanID:         u.planID,
		CustomerID:     input.CustomerID,
		CardID:         input.CardID,
		StartDate:      u.now(),
		IdempotencyKey: fincodeIdempotencyKey("subscription:" + input.CustomerID + ":" + input.CardID),
	})
	if err != nil {
		return err
	}

	subsc, ok := subscriptionEnabled(subscription.Status)
	if !ok {
		subsc = false
	}
	return u.repo.UpdateFincodeSubscription(user.ID(), subscription.ID, subsc)
}
