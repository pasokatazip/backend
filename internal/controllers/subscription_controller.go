package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/pasokatazip/backend/internal/controllers/dto"
	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/infrastructure/middleware"
	"github.com/pasokatazip/backend/internal/usecases"
)

type StartSubscriptionUsecase interface {
	Execute(ctx context.Context, userID domain.UserID) (domain.FincodeCardSession, error)
}

type CancelSubscriptionUsecase interface {
	Execute(ctx context.Context, userID domain.UserID) error
}

type GetSubscriptionUsecase interface {
	Execute(userID domain.UserID) (usecases.FincodeSubscriptionStatus, error)
}

type SubscriptionController struct {
	start  StartSubscriptionUsecase
	cancel CancelSubscriptionUsecase
	get    GetSubscriptionUsecase
}

func NewSubscriptionController(
	start StartSubscriptionUsecase,
	cancel CancelSubscriptionUsecase,
	get GetSubscriptionUsecase,
) *SubscriptionController {
	return &SubscriptionController{start: start, cancel: cancel, get: get}
}

func (c *SubscriptionController) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(w, r)
	if !ok {
		return
	}

	status, err := c.get.Execute(userID)
	if err != nil {
		writeSubscriptionError(w, err, "failed to get subscription")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewFincodeSubscriptionStatusResponse(status))
}

func (c *SubscriptionController) Start(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(w, r)
	if !ok {
		return
	}

	session, err := c.start.Execute(r.Context(), userID)
	if err != nil {
		writeSubscriptionError(w, err, "failed to start subscription")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.NewFincodeCheckoutResponse(session))
}

func (c *SubscriptionController) Cancel(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(w, r)
	if !ok {
		return
	}

	if err := c.cancel.Execute(r.Context(), userID); err != nil {
		writeSubscriptionError(w, err, "failed to cancel subscription")
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func authenticatedUserID(w http.ResponseWriter, r *http.Request) (domain.UserID, bool) {
	value, ok := middleware.GetUserID(r.Context())
	if !ok || !domain.IsValidUserID(domain.UserID(value)) {
		http.Error(w, domain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return "", false
	}
	return domain.UserID(value), true
}

func writeSubscriptionError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, domain.ErrValidation):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, domain.ErrNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, domain.ErrAlreadyExists):
		http.Error(w, err.Error(), http.StatusConflict)
	default:
		http.Error(w, fallback, http.StatusBadGateway)
	}
}
