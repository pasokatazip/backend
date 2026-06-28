package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type ensureCustomerRepository struct {
	domain.UserRepository
	user              domain.User
	findErr           error
	updatedUserID     domain.UserID
	updatedCustomerID string
	updateErr         error
}

func (r *ensureCustomerRepository) FindByID(domain.UserID) (domain.User, error) {
	return r.user, r.findErr
}

func (r *ensureCustomerRepository) UpdateFincodeCustomerID(userID domain.UserID, customerID string) error {
	r.updatedUserID = userID
	r.updatedCustomerID = customerID
	return r.updateErr
}

type ensureCustomerGateway struct {
	domain.FincodeGateway
	input  domain.FincodeCustomerInput
	result domain.FincodeCustomer
	err    error
	calls  int
}

func (g *ensureCustomerGateway) CreateCustomer(
	_ context.Context,
	input domain.FincodeCustomerInput,
) (domain.FincodeCustomer, error) {
	g.calls++
	g.input = input
	return g.result, g.err
}

func TestEnsureFincodeCustomerCreatesAndSavesCustomer(t *testing.T) {
	t.Parallel()

	userID := domain.UserID("4c0c926e-6a13-4bf4-8ae4-593c4047280f")
	repo := &ensureCustomerRepository{
		user: domain.NewUser(userID, "user@example.com", "hash", false, nil, nil, time.Time{}),
	}
	gateway := &ensureCustomerGateway{
		result: domain.FincodeCustomer{ID: string(userID), Email: "user@example.com"},
	}
	usecase := NewEnsureFincodeCustomer(repo, gateway)

	customer, err := usecase.Execute(context.Background(), userID)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if customer.ID != string(userID) {
		t.Errorf("customer ID = %q", customer.ID)
	}
	if gateway.calls != 1 {
		t.Errorf("gateway calls = %d, want 1", gateway.calls)
	}
	if gateway.input.ID != string(userID) || gateway.input.IdempotencyKey != string(userID) {
		t.Errorf("gateway input = %+v", gateway.input)
	}
	if repo.updatedUserID != userID || repo.updatedCustomerID != string(userID) {
		t.Errorf("repository update = (%q, %q)", repo.updatedUserID, repo.updatedCustomerID)
	}
}

func TestEnsureFincodeCustomerReusesExistingCustomer(t *testing.T) {
	t.Parallel()

	userID := domain.UserID("4c0c926e-6a13-4bf4-8ae4-593c4047280f")
	customerID := "c_existing"
	repo := &ensureCustomerRepository{
		user: domain.NewUser(
			userID,
			"user@example.com",
			"hash",
			false,
			&customerID,
			nil,
			time.Time{},
		),
	}
	gateway := &ensureCustomerGateway{}
	usecase := NewEnsureFincodeCustomer(repo, gateway)

	customer, err := usecase.Execute(context.Background(), userID)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if customer.ID != customerID {
		t.Errorf("customer ID = %q, want %q", customer.ID, customerID)
	}
	if gateway.calls != 0 {
		t.Errorf("gateway calls = %d, want 0", gateway.calls)
	}
}

func TestEnsureFincodeCustomerDoesNotSaveGatewayFailure(t *testing.T) {
	t.Parallel()

	userID := domain.UserID("4c0c926e-6a13-4bf4-8ae4-593c4047280f")
	repo := &ensureCustomerRepository{
		user: domain.NewUser(userID, "user@example.com", "hash", false, nil, nil, time.Time{}),
	}
	wantErr := errors.New("fincode unavailable")
	gateway := &ensureCustomerGateway{err: wantErr}
	usecase := NewEnsureFincodeCustomer(repo, gateway)

	_, err := usecase.Execute(context.Background(), userID)
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}
	if repo.updatedCustomerID != "" {
		t.Errorf("customer ID was saved after gateway failure: %q", repo.updatedCustomerID)
	}
}
