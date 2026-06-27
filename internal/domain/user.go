package domain

import (
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

func (u User) CreatedAt() time.Time {
	return u.createdAt
}

type UserRepository interface {
	Create(user User) (User, error)
	FindByEmail(email string) (User, error)
	FindByID(id UserID) (User, error)
	FindByFincodeCustomerID(customerID string) (User, error)
	FindByFincodeSubscriptionID(subscriptionID string) (User, error)
	UpdateFincodeCustomerID(id UserID, customerID string) error
	UpdateFincodeSubscription(id UserID, subscriptionID string, subsc bool) error
	UpdateSubscriptionStatus(id UserID, subsc bool) error
}
