package middleware

import (
	"net/http"
	"strings"
)

const (
	corsAllowedMethods = "GET, POST, PUT, PATCH, DELETE, OPTIONS"
	corsAllowedHeaders = "Authorization, Content-Type"
	corsMaxAge         = "86400"
)

func CORS(allowedOriginsCSV string) Middleware {
	allowedOrigins, allowAnyOrigin := parseAllowedOrigins(allowedOriginsCSV)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && isAllowedOrigin(origin, allowedOrigins, allowAnyOrigin) {
				if allowAnyOrigin {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Add("Vary", "Origin")
				}
				w.Header().Set("Access-Control-Allow-Methods", corsAllowedMethods)
				w.Header().Set("Access-Control-Allow-Headers", corsAllowedHeaders)
				w.Header().Set("Access-Control-Max-Age", corsMaxAge)
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func parseAllowedOrigins(allowedOriginsCSV string) (map[string]struct{}, bool) {
	allowedOrigins := make(map[string]struct{})
	allowAnyOrigin := false

	for _, origin := range strings.Split(allowedOriginsCSV, ",") {
		origin = strings.TrimSpace(origin)
		if origin == "" {
			continue
		}
		if origin == "*" {
			allowAnyOrigin = true
			continue
		}
		allowedOrigins[origin] = struct{}{}
	}

	return allowedOrigins, allowAnyOrigin
}

func isAllowedOrigin(origin string, allowedOrigins map[string]struct{}, allowAnyOrigin bool) bool {
	if allowAnyOrigin {
		return true
	}
	_, ok := allowedOrigins[origin]
	return ok
}
