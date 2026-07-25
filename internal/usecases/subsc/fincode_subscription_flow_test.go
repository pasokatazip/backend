package subsc

import (
	"context"
	"testing"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/usecases"
)

type subscriptionFlowRepository struct {
	domain.UserRepository
	user                  domain.User
	updatedSubscriptionID string
	updatedSubsc          bool
}

func (r *subscriptionFlowRepository) FindByID(domain.UserID) (domain.User, error) {
	return r.user, nil
}

func (r *subscriptionFlowRepository) FindByFincodeCustomerID(string) (domain.User, error) {
	return r.user, nil
}

func (r *subscriptionFlowRepository) UpdateFincodeSubscription(
	_ domain.UserID,
	subscriptionID string,
	subsc bool,
) error {
	r.updatedSubscriptionID = subscriptionID
	r.updatedSubsc = subsc
	return nil
}

type subscriptionFlowGateway struct {
	domain.FincodeGateway
	cardSessionInput  domain.FincodeCardSessionInput
	cardSession       domain.FincodeCardSession
	subscriptionInput domain.FincodeSubscriptionInput
	subscription      domain.FincodeSubscription
	canceledID        string
}

func (g *subscriptionFlowGateway) CreateCardSession(
	_ context.Context,
	input domain.FincodeCardSessionInput,
) (domain.FincodeCardSession, error) {
	g.cardSessionInput = input
	return g.cardSession, nil
}

func (g *subscriptionFlowGateway) CreateSubscription(
	_ context.Context,
	input domain.FincodeSubscriptionInput,
) (domain.FincodeSubscription, error) {
	g.subscriptionInput = input
	return g.subscription, nil
}

func (g *subscriptionFlowGateway) CancelSubscription(_ context.Context, subscriptionID string) error {
	g.canceledID = subscriptionID
	return nil
}

type customerEnsurerStub struct {
	customer domain.FincodeCustomer
}

func (s customerEnsurerStub) Execute(context.Context, domain.UserID) (domain.FincodeCustomer, error) {
	return s.customer, nil
}

func TestStartFincodeSubscriptionCreatesCardSession(t *testing.T) {
	t.Parallel()

	userID := domain.UserID("4c0c926e-6a13-4bf4-8ae4-593c4047280f")
	gateway := &subscriptionFlowGateway{
		cardSession: domain.FincodeCardSession{ID: "session-id", LinkURL: "https://example.com/checkout"},
	}
	usecase := NewStartFincodeSubscription(
		&subscriptionFlowRepository{
			user: domain.NewUser(userID, "user@example.com", "hash", false, nil, nil, time.Time{}),
		},
		customerEnsurerStub{customer: domain.FincodeCustomer{ID: "customer-id"}},
		gateway,
		"Example",
		30*time.Minute,
	)

	session, err := usecase.Execute(context.Background(), userID)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if session.ID != "session-id" || gateway.cardSessionInput.CustomerID != "customer-id" {
		t.Errorf("session/input = %+v / %+v", session, gateway.cardSessionInput)
	}
	if gateway.cardSessionInput.ExpiresAt.IsZero() {
		t.Error("session expiry was not set")
	}
}

func TestCardRegistrationCreatesSubscription(t *testing.T) {
	t.Parallel()

	userID := domain.UserID("4c0c926e-6a13-4bf4-8ae4-593c4047280f")
	customerID := "customer-id"
	repo := &subscriptionFlowRepository{
		user: domain.NewUser(userID, "user@example.com", "hash", false, &customerID, nil, time.Time{}),
	}
	gateway := &subscriptionFlowGateway{
		subscription: domain.FincodeSubscription{ID: "subscription-id", Status: "ACTIVE"},
	}
	usecase := NewCardRegistration(repo, gateway, "plan-id")
	usecase.now = func() time.Time {
		return time.Date(2026, 6, 27, 0, 0, 0, 0, time.UTC)
	}

	err := usecase.Execute(context.Background(), usecases.CardRegistrationInput{
		CustomerID: customerID,
		CardID:     "card-id",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if gateway.subscriptionInput.PlanID != "plan-id" || gateway.subscriptionInput.CardID != "card-id" {
		t.Errorf("subscription input = %+v", gateway.subscriptionInput)
	}
	if gateway.subscriptionInput.IdempotencyKey == "" {
		t.Error("idempotency key was not set")
	}
	if repo.updatedSubscriptionID != "subscription-id" || !repo.updatedSubsc {
		t.Errorf("repository update = (%q, %v)", repo.updatedSubscriptionID, repo.updatedSubsc)
	}
}

func TestCancelFincodeSubscriptionCallsGateway(t *testing.T) {
	t.Parallel()

	userID := domain.UserID("4c0c926e-6a13-4bf4-8ae4-593c4047280f")
	subscriptionID := "subscription-id"
	repo := &subscriptionFlowRepository{
		user: domain.NewUser(userID, "user@example.com", "hash", true, nil, &subscriptionID, time.Time{}),
	}
	gateway := &subscriptionFlowGateway{}
	usecase := NewCancelFincodeSubscription(repo, gateway)

	if err := usecase.Execute(context.Background(), userID); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if gateway.canceledID != subscriptionID {
		t.Errorf("canceled ID = %q", gateway.canceledID)
	}
}
