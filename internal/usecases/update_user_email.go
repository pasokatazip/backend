package usecases

import (
	"errors"
	"strings"

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
	input.CurrentEmail = strings.TrimSpace(input.CurrentEmail)
	input.NewEmail = strings.TrimSpace(input.NewEmail)
	if input.CurrentEmail == "" || input.CurrentPassword == "" || input.NewEmail == "" || u.hasher == nil {
		return domain.ErrValidation
	}

	user, err := u.repo.FindByEmail(input.CurrentEmail)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrUnauthorized
		}
		return err
	}
	if err := u.hasher.Compare(user.Password(), input.CurrentPassword); err != nil {
		return domain.ErrUnauthorized
	}
	if strings.EqualFold(user.Email(), input.NewEmail) {
		return domain.ErrValidation
	}

	if _, err := u.repo.FindByEmail(input.NewEmail); err == nil {
		return domain.ErrAlreadyExists
	} else if !errors.Is(err, domain.ErrNotFound) {
		return err
	}

	return u.repo.UpdateEmail(user.ID(), input.NewEmail)
}
