package onetime

import (
	"context"
	"testing"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type purchaseRepository struct {
	domain.UserRepository
	user      domain.User
	billingID string
	entitled  bool
}

func (r *purchaseRepository) FindByID(domain.UserID) (domain.User, error) {
	return r.user, nil
}

func (r *purchaseRepository) UpdateFincodeBilling(_ domain.UserID, id string, entitled bool) error {
	r.billingID, r.entitled = id, entitled
	return nil
}

type cardGateway struct {
	cards []domain.FincodeCard
}

func (g cardGateway) ListCards(context.Context, string) ([]domain.FincodeCard, error) {
	return g.cards, nil
}

type paymentGateway struct {
	existing     domain.FincodePayment
	getErr       error
	createInput  domain.FincodePaymentInput
	executeInput domain.FincodePaymentInput
}

func (g *paymentGateway) GetPayment(context.Context, string) (domain.FincodePayment, error) {
	return g.existing, g.getErr
}

func (g *paymentGateway) CreatePayment(
	_ context.Context,
	input domain.FincodePaymentInput,
) (domain.FincodePayment, error) {
	g.createInput = input
	return domain.FincodePayment{
		ID: input.ID, AccessID: "access-id", Status: "UNPROCESSED",
	}, nil
}

func (g *paymentGateway) ExecutePayment(
	_ context.Context,
	input domain.FincodePaymentInput,
) (domain.FincodePayment, error) {
	g.executeInput = input
	return domain.FincodePayment{
		ID: input.ID, AccessID: input.AccessID, Status: "CAPTURED",
	}, nil
}

func TestConfirmPurchaseChargesDefaultCardAndReturnsPurchased(t *testing.T) {
	t.Parallel()
	userID := domain.UserID("4c0c926e-6a13-4bf4-8ae4-593c4047280f")
	customerID := "customer-id"
	repo := &purchaseRepository{
		user: domain.NewUser(userID, "user@example.com", "hash", false, &customerID, nil, time.Time{}),
	}
	payments := &paymentGateway{getErr: domain.ErrNotFound}
	uc := NewConfirmFincodePurchase(
		repo,
		cardGateway{cards: []domain.FincodeCard{
			{ID: "other-card", DefaultFlag: "0"},
			{ID: "default-card", DefaultFlag: "1"},
		}},
		payments,
		999,
	)

	status, err := uc.Execute(context.Background(), userID)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if payments.createInput.Amount != 999 {
		t.Errorf("amount = %d", payments.createInput.Amount)
	}
	if payments.executeInput.CardID != "default-card" ||
		payments.executeInput.AccessID != "access-id" {
		t.Errorf("execute input = %+v", payments.executeInput)
	}
	if !status.Purchased || !repo.entitled || repo.billingID == "" {
		t.Errorf("status=%+v repo=(%q,%v)", status, repo.billingID, repo.entitled)
	}
}

func TestConfirmPurchaseIsIdempotentForPurchasedUser(t *testing.T) {
	t.Parallel()
	userID := domain.UserID("4c0c926e-6a13-4bf4-8ae4-593c4047280f")
	customerID := "customer-id"
	paymentID := "payment-id"
	repo := &purchaseRepository{
		user: domain.NewUser(userID, "user@example.com", "hash", true, &customerID, &paymentID, time.Time{}),
	}
	payments := &paymentGateway{}
	uc := NewConfirmFincodePurchase(repo, cardGateway{}, payments, 999)

	status, err := uc.Execute(context.Background(), userID)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !status.Purchased || payments.createInput.ID != "" {
		t.Errorf("status=%+v create=%+v", status, payments.createInput)
	}
}
