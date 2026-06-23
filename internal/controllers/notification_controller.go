package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/pasokatazip/backend/internal/controllers/dto"
	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/infrastructure/middleware"
	"github.com/pasokatazip/backend/internal/presenter"
	"github.com/pasokatazip/backend/internal/usecases"
)

type NotificationController struct {
	createNotification       *usecases.CreateNotification
	updateNotification       *usecases.UpdateNotification
	findNotificationByUserID *usecases.FindNotificationByUserID
}

func NewNotificationController(
	createNotification *usecases.CreateNotification,
	updateNotification *usecases.UpdateNotification,
	findNotificationByUserID *usecases.FindNotificationByUserID,
) *NotificationController {
	return &NotificationController{
		createNotification:       createNotification,
		updateNotification:       updateNotification,
		findNotificationByUserID: findNotificationByUserID,
	}
}

func (c *NotificationController) Create(w http.ResponseWriter, r *http.Request) {
	req, userID, ok := c.requestInput(w, r)
	if !ok {
		return
	}

	notification, err := c.createNotification.Execute(req.ToUseCaseInput(userID))
	if err != nil {
		c.handleError(w, err, "failed to create notification")
		return
	}

	c.writeNotification(w, http.StatusCreated, notification)
}

func (c *NotificationController) Update(w http.ResponseWriter, r *http.Request) {
	req, userID, ok := c.requestInput(w, r)
	if !ok {
		return
	}

	notification, err := c.updateNotification.Execute(req.ToUseCaseInput(userID))
	if err != nil {
		c.handleError(w, err, "failed to update notification")
		return
	}

	c.writeNotification(w, http.StatusOK, notification)
}

func (c *NotificationController) FindByUserID(w http.ResponseWriter, r *http.Request) {
	userIDString, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, domain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	notification, err := c.findNotificationByUserID.Execute(domain.UserID(userIDString))
	if err != nil {
		c.handleError(w, err, "failed to fetch notification")
		return
	}

	c.writeNotification(w, http.StatusOK, notification)
}

func (c *NotificationController) requestInput(w http.ResponseWriter, r *http.Request) (dto.NotificationRequest, domain.UserID, bool) {
	var req dto.NotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return dto.NotificationRequest{}, "", false
	}

	userIDString, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, domain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return dto.NotificationRequest{}, "", false
	}

	return req, domain.UserID(userIDString), true
}

func (c *NotificationController) writeNotification(w http.ResponseWriter, status int, notification domain.Notification) {
	pr := presenter.NewNotificationPresenter()
	output := pr.Output(notification)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(dto.NewNotificationResponse(output))
}

func (c *NotificationController) handleError(w http.ResponseWriter, err error, fallback string) {
	if errors.Is(err, domain.ErrValidation) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if errors.Is(err, domain.ErrNotFound) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	http.Error(w, fallback, http.StatusInternalServerError)
}
