package usecases

import (
	"time"

	"github.com/pasokatazip/backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type TokenGenerator interface {
	Generate(user domain.User) (string, time.Time, error)
}

type LoginInput struct {
	Email    string
	Password string
}

type Login struct {
	repo     domain.UserRepository
	tokenGen TokenGenerator
}

func NewLogin(repo domain.UserRepository, tokenGen TokenGenerator) *Login {
	return &Login{repo: repo, tokenGen: tokenGen}
}

func (l *Login) Execute(input LoginInput) (string, time.Time, domain.User, error) {
	user, err := l.repo.FindByEmail(input.Email)
	if err != nil {
		return "", time.Time{}, domain.User{}, domain.ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password()), []byte(input.Password)); err != nil {
		return "", time.Time{}, domain.User{}, domain.ErrUnauthorized
	}

	token, expiresAt, err := l.tokenGen.Generate(user)
	if err != nil {
		return "", time.Time{}, domain.User{}, err
	}

	return token, expiresAt, user, nil
}
