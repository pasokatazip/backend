package controllers

import (
	"errors"
	"net/http"

	"github.com/pasokatazip/backend/internal/domain"
)

func writeDomainError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, domain.ErrValidation):
		http.Error(w, domain.ErrValidation.Error(), http.StatusBadRequest)
	case errors.Is(err, domain.ErrUnauthorized):
		http.Error(w, domain.ErrUnauthorized.Error(), http.StatusUnauthorized)
	case errors.Is(err, domain.ErrNotFound):
		http.Error(w, domain.ErrNotFound.Error(), http.StatusNotFound)
	case errors.Is(err, domain.ErrAlreadyExists):
		http.Error(w, domain.ErrAlreadyExists.Error(), http.StatusConflict)
	case errors.Is(err, domain.ErrSubscriptionRequired):
		http.Error(w, domain.ErrSubscriptionRequired.Error(), http.StatusForbidden)
	case errors.Is(err, domain.ErrExternalService):
		http.Error(w, domain.ErrExternalService.Error(), http.StatusBadGateway)
	case errors.Is(err, domain.ErrInternal):
		http.Error(w, fallback, http.StatusInternalServerError)
	default:
		http.Error(w, fallback, http.StatusInternalServerError)
	}
}
