package domain

import (
	"context"
	"time"
)

type User struct {
	id                    UserID
	email                 string
	password              string
	subsc                 bool
	fincodeCustomerID     *string
	fincodeSubscriptionID *string
	createdAt             time.Time
}

func NewUser(
	id UserID,
	email string,
	password string,
	subsc bool,
	fincodeCustomerID *string,
	fincodeSubscriptionID *string,
	createdAt time.Time,
) User {
	return User{
		id:                    id,
		email:                 email,
		password:              password,
		subsc:                 subsc,
		fincodeCustomerID:     fincodeCustomerID,
		fincodeSubscriptionID: fincodeSubscriptionID,
		createdAt:             createdAt,
	}
}

func (u User) ID() UserID {
	return u.id
}

func (u User) Email() string {
	return u.email
}

func (u User) Password() string {
	return u.password
}

func (u User) Subsc() bool {
	return u.subsc
}

func (u User) FincodeCustomerID() *string {
	return u.fincodeCustomerID
}

func (u User) FincodeSubscriptionID() *string {
	return u.fincodeSubscriptionID
}

// FincodeBillingID is the provider transaction ID for the configured billing
// mode. The underlying legacy column is retained to keep existing data usable.
func (u User) FincodeBillingID() *string {
	return u.fincodeSubscriptionID
}

func (u User) CreatedAt() time.Time {
	return u.createdAt
}

type UserRepository interface {
	Create(ctx context.Context, user User) (User, error)
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByID(ctx context.Context, id UserID) (User, error)
	FindByFincodeCustomerID(ctx context.Context, customerID string) (User, error)
	FindByFincodeSubscriptionID(ctx context.Context, subscriptionID string) (User, error)
	UpdateFincodeCustomerID(ctx context.Context, id UserID, customerID string) error
	UpdateFincodeSubscription(ctx context.Context, id UserID, subscriptionID string, subsc bool) error
	UpdateFincodeBilling(ctx context.Context, id UserID, billingID string, entitled bool) error
	UpdateSubscriptionStatus(ctx context.Context, id UserID, subsc bool) error
	UpdateEmail(ctx context.Context, id UserID, email string) error
	UpdatePassword(ctx context.Context, id UserID, password string) error
}
