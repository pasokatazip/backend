package subsc

import (
	"context"

	"github.com/pasokatazip/backend/internal/domain"
)

type CancelFincodeSubscription struct {
	repo    domain.UserRepository
	gateway domain.FincodeSubscriptionGateway
}

func NewCancelFincodeSubscription(
	repo domain.UserRepository,
	gateway domain.FincodeSubscriptionGateway,
) *CancelFincodeSubscription {
	return &CancelFincodeSubscription{repo: repo, gateway: gateway}
}

func (u *CancelFincodeSubscription) Execute(ctx context.Context, userID domain.UserID) error {
	if !domain.IsValidUserID(userID) || u.gateway == nil {
		return domain.ErrValidation
	}

	user, err := u.repo.FindByID(userID)
	if err != nil {
		return err
	}
	if user.FincodeSubscriptionID() == nil || *user.FincodeSubscriptionID() == "" {
		return domain.ErrNotFound
	}

	// subsc is finalized by subscription.card.delete Webhook.
	return u.gateway.CancelSubscription(ctx, *user.FincodeSubscriptionID())
}
