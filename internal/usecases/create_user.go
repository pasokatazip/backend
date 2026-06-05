package usecases

import (
	"time"

	"github.com/pasokatazip/backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
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
	repo domain.UserRepository
}

func NewCreateUser(repo domain.UserRepository) *CreateUser {
	return &CreateUser{repo: repo}
}

func (u *CreateUser) Execute(input CreateUserInput) (domain.User, error) {
	if input.Email == "" || input.Password == "" {
		return domain.User{}, domain.ErrValidation
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, err
	}

	// default subscription flag is false on creation
	newUser := domain.NewUser(domain.NewUserID(), input.Email, string(hashedPassword), false, time.Now().UTC())

	savedUser, err := u.repo.Create(newUser)
	if err != nil {
		return domain.User{}, err
	}

	return savedUser, nil
}
