package middleware

import (
	"net/http"

	"github.com/pasokatazip/backend/internal/domain"
)

func RequireSubscMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		subsc, ok := GetSubsc(r.Context())
		if !ok || !subsc {
			http.Error(
				w,
				domain.ErrSubscriptionRequired.Error(),
				http.StatusForbidden,
			)
			return
		}

		next.ServeHTTP(w, r)
	})
}
