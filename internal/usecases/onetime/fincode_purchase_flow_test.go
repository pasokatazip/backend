package onetime

import (
	"context"
	"testing"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/usecases"
)

type purchaseRepository struct {
	domain.UserRepository
	user      domain.User
	billingID string
	entitled  bool
}

func (r *purchaseRepository) FindByFincodeCustomerID(string) (domain.User, error) {
	return r.user, nil
}

func (r *purchaseRepository) UpdateFincodeBilling(_ domain.UserID, id string, entitled bool) error {
	r.billingID, r.entitled = id, entitled
	return nil
}

type paymentGateway struct {
	createInput  domain.FincodePaymentInput
	executeInput domain.FincodePaymentInput
}

func (g *paymentGateway) CreatePayment(_ context.Context, input domain.FincodePaymentInput) (domain.FincodePayment, error) {
	g.createInput = input
	return domain.FincodePayment{ID: input.ID, Status: "UNPROCESSED"}, nil
}

func (g *paymentGateway) ExecutePayment(_ context.Context, input domain.FincodePaymentInput) (domain.FincodePayment, error) {
	g.executeInput = input
	return domain.FincodePayment{ID: input.ID, Status: "CAPTURED"}, nil
}

func TestCardRegistrationExecutesOneTimePayment(t *testing.T) {
	t.Parallel()
	userID := domain.UserID("4c0c926e-6a13-4bf4-8ae4-593c4047280f")
	customerID := "customer-id"
	repo := &purchaseRepository{
		user: domain.NewUser(userID, "user@example.com", "hash", false, &customerID, nil, time.Time{}),
	}
	gateway := &paymentGateway{}
	uc := NewCardRegistration(repo, gateway, 1980)

	err := uc.Execute(context.Background(), usecases.CardRegistrationInput{
		CustomerID: customerID,
		CardID:     "card-id",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if gateway.createInput.Amount != 1980 || gateway.createInput.ID == "" {
		t.Errorf("create input = %+v", gateway.createInput)
	}
	if gateway.executeInput.CustomerID != customerID || gateway.executeInput.CardID != "card-id" {
		t.Errorf("execute input = %+v", gateway.executeInput)
	}
	if repo.billingID == "" || !repo.entitled {
		t.Errorf("billing update = (%q, %v)", repo.billingID, repo.entitled)
	}
}

func TestCardRegistrationDoesNotChargeEntitledUserAgain(t *testing.T) {
	t.Parallel()
	userID := domain.UserID("4c0c926e-6a13-4bf4-8ae4-593c4047280f")
	customerID := "customer-id"
	paymentID := "payment-id"
	repo := &purchaseRepository{
		user: domain.NewUser(userID, "user@example.com", "hash", true, &customerID, &paymentID, time.Time{}),
	}
	gateway := &paymentGateway{}
	uc := NewCardRegistration(repo, gateway, 1980)

	if err := uc.Execute(context.Background(), usecases.CardRegistrationInput{
		CustomerID: customerID, CardID: "card-id",
	}); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if gateway.createInput.ID != "" {
		t.Fatal("already entitled user was charged")
	}
}
