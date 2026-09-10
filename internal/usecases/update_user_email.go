package usecases

import (
	"errors"

	"github.com/pasokatazip/backend/internal/domain"
)

type UpdateUserEmailInput struct {
	CurrentEmail    string
	CurrentPassword string
	NewEmail        string
}

type UpdateUserEmail struct {
	repo   domain.UserRepository
	hasher PasswordHasher
}

func NewUpdateUserEmail(repo domain.UserRepository, hasher PasswordHasher) *UpdateUserEmail {
	return &UpdateUserEmail{repo: repo, hasher: hasher}
}

func (u *UpdateUserEmail) Execute(input UpdateUserEmailInput) error {
	currentEmail, currentEmailOK := normalizeAndValidateEmail(input.CurrentEmail)
	newEmail, newEmailOK := normalizeAndValidateEmail(input.NewEmail)
	if !currentEmailOK || !newEmailOK || !isValidPassword(input.CurrentPassword) || u.hasher == nil {
		return domain.ErrValidation
	}

	user, err := u.repo.FindByEmail(currentEmail)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrUnauthorized
		}
		return err
	}
	if err := u.hasher.Compare(user.Password(), input.CurrentPassword); err != nil {
		return domain.ErrUnauthorized
	}
	if user.Email() == newEmail {
		return domain.ErrValidation
	}

	if _, err := u.repo.FindByEmail(newEmail); err == nil {
		return domain.ErrAlreadyExists
	} else if !errors.Is(err, domain.ErrNotFound) {
		return err
	}

	return u.repo.UpdateEmail(user.ID(), newEmail)
}
