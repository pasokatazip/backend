package auth

import (
	"fmt"

	"github.com/pasokatazip/backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type BCryptPasswordHasher struct{}

func NewBCryptPasswordHasher() *BCryptPasswordHasher {
	return &BCryptPasswordHasher{}
}

func (h *BCryptPasswordHasher) Hash(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("%w: hash password: %v", domain.ErrInternal, err)
	}
	return string(b), nil
}

func (h *BCryptPasswordHasher) Compare(hash string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
