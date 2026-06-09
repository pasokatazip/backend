package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	jwt "github.com/golang-jwt/jwt/v5"

	"github.com/pasokatazip/backend/internal/domain"
)

type contextKey string

const (
	contextKeyUserID contextKey = "userID"
	contextKeySubsc  contextKey = "subsc"
)

func AuthMiddleware(next http.Handler) http.Handler {
	secret := os.Getenv("JWT_SECRET")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, domain.ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}

		parts := strings.Fields(auth)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, domain.ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %T", t.Method)
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, domain.ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, domain.ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		if uid, ok := claims["user_id"].(string); ok {
			ctx = context.WithValue(ctx, contextKeyUserID, uid)
		}
		if s, ok := claims["subsc"].(bool); ok {
			ctx = context.WithValue(ctx, contextKeySubsc, s)
		} else if f, ok := claims["subsc"].(float64); ok {
			ctx = context.WithValue(ctx, contextKeySubsc, f != 0)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(contextKeyUserID).(string)
	return v, ok
}

func GetSubsc(ctx context.Context) (bool, bool) {
	v := ctx.Value(contextKeySubsc)
	if v == nil {
		return false, false
	}
	b, ok := v.(bool)
	return b, ok
}
