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
		"subsc":   user.Subsc(),
		"iat":     jwt.NewNumericDate(now),
		"exp":     jwt.NewNumericDate(expiry),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(g.secret)
	if err != nil {
		return "", time.Time{}, err
	}

	return signed, expiry, nil
}

func (g *JWTTokenGenerator) Parse(tokenString string) (domain.UserID, time.Time, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return g.secret, nil
	})
	if err != nil || !token.Valid {
		return "", time.Time{}, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", time.Time{}, jwt.ErrInvalidKey
	}

	uid, ok := claims["user_id"].(string)
	if !ok {
		return "", time.Time{}, jwt.ErrInvalidKey
	}

	var expiresAt time.Time
	if exp, ok := claims["exp"].(jwt.NumericDate); ok {
		expiresAt = exp.Time
	} else if f, ok := claims["exp"].(float64); ok {
		expiresAt = time.Unix(int64(f), 0)
	}

	return domain.UserID(uid), expiresAt, nil
}
