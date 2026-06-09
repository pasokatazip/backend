package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pasokatazip/backend/internal/domain"
)

type JWTTokenGenerator struct {
	secret        []byte
	expiryMinutes int
}

func NewJWTTokenGenerator(secret string, expiryMinutes int) *JWTTokenGenerator {
	return &JWTTokenGenerator{secret: []byte(secret), expiryMinutes: expiryMinutes}
}

func (g *JWTTokenGenerator) Generate(user domain.User) (string, time.Time, error) {
	now := time.Now().UTC()
	expiry := now.Add(time.Duration(g.expiryMinutes) * time.Minute)

	claims := jwt.MapClaims{
		"user_id": string(user.ID()),
		"subsc":    user.Subsc(),
		"iat":      jwt.NewNumericDate(now),
		"exp":      jwt.NewNumericDate(expiry),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(g.secret)
	if err != nil {
		return "", time.Time{}, err
	}

	return signed, expiry, nil
}
