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

type FincodeGateway interface {
	CreateCustomer(ctx context.Context, input FincodeCustomerInput) (FincodeCustomer, error)
	CreateCardSession(ctx context.Context, input FincodeCardSessionInput) (FincodeCardSession, error)
	CreateSubscription(ctx context.Context, input FincodeSubscriptionInput) (FincodeSubscription, error)
	CancelSubscription(ctx context.Context, subscriptionID string) error
}
