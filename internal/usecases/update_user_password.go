package usecases

import (
	"errors"

	"github.com/pasokatazip/backend/internal/domain"
)

type UpdateUserPasswordInput struct {
	Email           string
	CurrentPassword string
	NewPassword     string
}

type UpdateUserPassword struct {
	repo   domain.UserRepository
	hasher PasswordHasher
}

func NewUpdateUserPassword(repo domain.UserRepository, hasher PasswordHasher) *UpdateUserPassword {
	return &UpdateUserPassword{repo: repo, hasher: hasher}
}

func (u *UpdateUserPassword) Execute(input UpdateUserPasswordInput) error {
	email, emailOK := normalizeAndValidateEmail(input.Email)
	if !emailOK ||
		!isValidPassword(input.CurrentPassword) ||
		!isValidPassword(input.NewPassword) ||
		input.CurrentPassword == input.NewPassword ||
		u.hasher == nil {
		return domain.ErrValidation
	}

	user, err := u.repo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrUnauthorized
		}
		return err
	}
	if err := u.hasher.Compare(user.Password(), input.CurrentPassword); err != nil {
		return domain.ErrUnauthorized
	}

	hashedPassword, err := u.hasher.Hash(input.NewPassword)
	if err != nil {
		return err
	}
	return u.repo.UpdatePassword(user.ID(), hashedPassword)
}
