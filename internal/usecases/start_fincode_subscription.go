package usecases

import (
	"context"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/timeutil"
)

type FincodeCustomerEnsurer interface {
	Execute(ctx context.Context, userID domain.UserID) (domain.FincodeCustomer, error)
}

type StartFincodeSubscription struct {
	repo           domain.UserRepository
	ensureCustomer FincodeCustomerEnsurer
	gateway        domain.FincodeGateway
	serviceName    string
	sessionTTL     time.Duration
}

func NewStartFincodeSubscription(
	repo domain.UserRepository,
	ensureCustomer FincodeCustomerEnsurer,
	gateway domain.FincodeGateway,
	serviceName string,
	sessionTTL time.Duration,
) *StartFincodeSubscription {
	return &StartFincodeSubscription{
		repo:           repo,
		ensureCustomer: ensureCustomer,
		gateway:        gateway,
		serviceName:    serviceName,
		sessionTTL:     sessionTTL,
	}
}

func (u *StartFincodeSubscription) Execute(
	ctx context.Context,
	userID domain.UserID,
) (domain.FincodeCardSession, error) {
	if !domain.IsValidUserID(userID) || u.repo == nil || u.ensureCustomer == nil || u.gateway == nil || u.sessionTTL <= 0 {
		return domain.FincodeCardSession{}, domain.ErrValidation
	}
	user, err := u.repo.FindByID(userID)
	if err != nil {
		return domain.FincodeCardSession{}, err
	}
	if user.Subsc() {
		return domain.FincodeCardSession{}, domain.ErrAlreadyExists
	}

	customer, err := u.ensureCustomer.Execute(ctx, userID)
	if err != nil {
		return domain.FincodeCardSession{}, err
	}

	return u.gateway.CreateCardSession(ctx, domain.FincodeCardSessionInput{
		CustomerID:      customer.ID,
		ShopServiceName: u.serviceName,
		ExpiresAt:       timeutil.NowJST().Add(u.sessionTTL),
	})
}
