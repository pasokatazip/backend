package onetime

import (
	"context"
	"errors"
	"strings"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/usecases"
)

type CardRegistration struct {
	repo    domain.UserRepository
	gateway domain.FincodePaymentGateway
	amount  int
}

func NewCardRegistration(repo domain.UserRepository, gateway domain.FincodePaymentGateway, amount int) *CardRegistration {
	return &CardRegistration{repo: repo, gateway: gateway, amount: amount}
}

func (u *CardRegistration) Execute(ctx context.Context, input usecases.CardRegistrationInput) error {
	if strings.TrimSpace(input.CustomerID) == "" || strings.TrimSpace(input.CardID) == "" || u.amount <= 0 || u.gateway == nil {
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
	if user.Subsc() {
		return nil
	}

	stableKey := usecases.FincodeIdempotencyKey("one-time:" + input.CustomerID)
	orderID := "ot_" + strings.ReplaceAll(stableKey, "-", "")[:27]

	// A previous webhook attempt may have registered the payment before its
	// response or execution failed. Resume that payment instead of reusing an
	// expired idempotency key or creating another charge.
	payment, getErr := u.gateway.GetPayment(ctx, orderID)
	if getErr != nil {
		payment, err = u.gateway.CreatePayment(ctx, domain.FincodePaymentInput{
			ID:             orderID,
			Amount:         u.amount,
			IdempotencyKey: domain.NewUUIDString(),
		})
		if err != nil {
			return err
		}
	}
	if paymentSucceeded(payment.Status) {
		return u.repo.UpdateFincodeBilling(user.ID(), payment.ID, true)
	}

	payment, err = u.gateway.ExecutePayment(ctx, domain.FincodePaymentInput{
		ID: payment.ID, CustomerID: input.CustomerID, CardID: input.CardID,
		IdempotencyKey: domain.NewUUIDString(),
	})
	if err != nil {
		return err
	}
	if !paymentSucceeded(payment.Status) {
		return domain.ErrExternalService
	}

	// The existing column/flag are retained for backward compatibility. In one-time
	// mode they represent payment ID and permanent entitlement, respectively.
	return u.repo.UpdateFincodeBilling(user.ID(), payment.ID, true)
}

func paymentSucceeded(status string) bool {
	switch strings.ToUpper(status) {
	case "CAPTURED", "SUCCEEDED":
		return true
	default:
		return false
	}
}
