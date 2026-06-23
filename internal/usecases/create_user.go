package usecases

import (
	"time"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/timeutil"
)

type CreateUserInput struct {
	Email    string
	Password string
}

type CreateUserOutput struct {
	ID        string
	Email     string
	Subsc     bool
	CreatedAt time.Time
}

type CreateUser struct {
	repo     domain.UserRepository
	tokenGen TokenGenerator
	hasher   PasswordHasher
}

func NewCreateUser(repo domain.UserRepository, tokenGen TokenGenerator, hasher PasswordHasher) *CreateUser {
	return &CreateUser{repo: repo, tokenGen: tokenGen, hasher: hasher}
}

func (u *CreateUser) Execute(input CreateUserInput) (domain.User, string, time.Time, error) {
	if input.Email == "" || input.Password == "" {
		return domain.User{}, "", time.Time{}, domain.ErrValidation
	}

	if u.hasher == nil {
		return domain.User{}, "", time.Time{}, domain.ErrValidation
	}
	hashedPassword, err := u.hasher.Hash(input.Password)
	if err != nil {
		return domain.User{}, "", time.Time{}, err
	}

	newUser := domain.NewUser(domain.NewUserID(), input.Email, string(hashedPassword), false, timeutil.NowJST())

	savedUser, err := u.repo.Create(newUser)
	if err != nil {
		return domain.User{}, "", time.Time{}, err
	}

	token, expiresAt, err := u.tokenGen.Generate(savedUser)
	if err != nil {
		return domain.User{}, "", time.Time{}, err
	}

	return savedUser, token, expiresAt, nil
}
