package domain

import (
	"context"
	"time"
)

type FincodeCustomerInput struct {
	ID             string
	Email          string
	IdempotencyKey string
}

type FincodeCustomer struct {
	ID    string
	Email string
}

type FincodeCardSessionInput struct {
	CustomerID      string
	ShopServiceName string
	ExpiresAt       time.Time
}

type FincodeCardSession struct {
	ID        string
	LinkURL   string
	ExpiresAt time.Time
}

type FincodeSubscriptionInput struct {
	PlanID         string
	CustomerID     string
	CardID         string
	StartDate      time.Time
	IdempotencyKey string
}

type FincodeSubscription struct {
	ID         string
	CustomerID string
	PlanID     string
	Status     string
}

type FincodeCustomerGateway interface {
	CreateCustomer(ctx context.Context, input FincodeCustomerInput) (FincodeCustomer, error)
}

type FincodeCardSessionGateway interface {
	CreateCardSession(ctx context.Context, input FincodeCardSessionInput) (FincodeCardSession, error)
}

type FincodeSubscriptionGateway interface {
	CreateSubscription(ctx context.Context, input FincodeSubscriptionInput) (FincodeSubscription, error)
	CancelSubscription(ctx context.Context, subscriptionID string) error
}

type FincodePaymentInput struct {
	ID             string
	Amount         int
	CustomerID     string
	CardID         string
	TransactionID  string
	IdempotencyKey string
}

type FincodePayment struct {
	ID            string
	TransactionID string
	Status        string
}

type FincodePaymentGateway interface {
	GetPayment(ctx context.Context, paymentID string) (FincodePayment, error)
	CreatePayment(ctx context.Context, input FincodePaymentInput) (FincodePayment, error)
	ExecutePayment(ctx context.Context, input FincodePaymentInput) (FincodePayment, error)
}

// FincodeGateway is kept as a convenience composite for the concrete fincode client.
// Use cases should depend on one of the smaller interfaces above.
type FincodeGateway interface {
	FincodeCustomerGateway
	FincodeCardSessionGateway
	FincodeSubscriptionGateway
	FincodePaymentGateway
}
