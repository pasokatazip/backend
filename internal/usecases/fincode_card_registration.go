package usecases

import (
	"errors"

	"github.com/pasokatazip/backend/internal/domain"
)

type CardRegistrationInput struct {
	CustomerID string
}

type CardRegistration struct {
	repo domain.UserRepository
}

func NewCardRegistration(repo domain.UserRepository) *CardRegistration {
	return &CardRegistration{repo: repo}
}

// Execute confirms that the fincode customer in the card webhook is associated
// with a local user. Normally the customer ID has already been saved before the
// card registration URL is issued. The UserID fallback supports a fincode
// customer created with the local UUID as its explicit customer ID.
func (u *CardRegistration) Execute(input CardRegistrationInput) error {
	if input.CustomerID == "" {
		return domain.ErrValidation
	}

	if _, err := u.repo.FindByFincodeCustomerID(input.CustomerID); err == nil {
		return nil
	} else if !errors.Is(err, domain.ErrNotFound) {
		return err
	} else if !domain.IsValidUserID(domain.UserID(input.CustomerID)) {
		return err
	}

	return u.repo.UpdateFincodeCustomerID(domain.UserID(input.CustomerID), input.CustomerID)
}
