package usecases

import (
	"fmt"
	"log"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type TokenGenerator interface {
	Generate(user domain.User) (string, time.Time, error)
}

type LoginInput struct {
	Email    string
	Password string
}

type Login struct {
	repo        domain.UserRepository
	tokenGen    TokenGenerator
	tokenParser TokenParser
	hasher      PasswordHasher
}

func NewLogin(repo domain.UserRepository, tokenGen TokenGenerator, tokenParser TokenParser, hasher PasswordHasher) *Login {
	return &Login{repo: repo, tokenGen: tokenGen, tokenParser: tokenParser, hasher: hasher}
}

func (l *Login) Execute(input LoginInput) (string, time.Time, domain.User, error) {
	user, err := l.repo.FindByEmail(input.Email)
	if err != nil {
		return "", time.Time{}, domain.User{}, domain.ErrUnauthorized
	}

	if l.hasher == nil {
		return "", time.Time{}, domain.User{}, domain.ErrUnauthorized
	}
	if err := l.hasher.Compare(user.Password(), input.Password); err != nil {
		return "", time.Time{}, domain.User{}, domain.ErrUnauthorized
	}

	token, expiresAt, err := l.tokenGen.Generate(user)
	if err != nil {
		return "", time.Time{}, domain.User{}, err
	}

	return token, expiresAt, user, nil
}

func (l *Login) ExecuteToken(tokenString string) (string, time.Time, domain.User, error) {
	if l.tokenParser == nil {
		log.Println("ExecuteToken: tokenParser is nil")
		return "", time.Time{}, domain.User{}, domain.ErrUnauthorized
	}

	uid, _, err := l.tokenParser.Parse(tokenString)
	if err != nil {
		log.Printf("ExecuteToken: token parse failed: %v\n", err)
		return "", time.Time{}, domain.User{}, fmt.Errorf("token parse failed: %w", err)
	}

	user, err := l.repo.FindByID(uid)
	if err != nil {
		log.Printf("ExecuteToken: user lookup failed for id %s: %v\n", uid, err)
		return "", time.Time{}, domain.User{}, fmt.Errorf("user not found: %w", err)
	}

	if l.tokenGen == nil {
		log.Println("ExecuteToken: tokenGen is nil")
		return "", time.Time{}, domain.User{}, domain.ErrUnauthorized
	}

	// The stored user is the source of truth for mutable claims such as subsc.
	// Issue a fresh token so a subscription update made by a webhook is reflected.
	refreshedToken, expiresAt, err := l.tokenGen.Generate(user)
	if err != nil {
		return "", time.Time{}, domain.User{}, fmt.Errorf("generate refreshed token: %w", err)
	}

	return refreshedToken, expiresAt, user, nil
}

type TokenParser interface {
	Parse(tokenString string) (domain.UserID, time.Time, error)
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash string, password string) error
}
