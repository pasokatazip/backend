package usecases

import (
	"encoding/json"

	"github.com/pasokatazip/backend/internal/domain"
)

type CreateNotification struct {
	repo domain.NotificationRepository
}

type NotificationInput struct {
	UserID           domain.UserID
	IsAllEnabled     bool
	IsYoyoEnabled    bool
	IsReportEnabled  bool
	IsMessageEnabled bool
	Subscription     json.RawMessage
}

type NotificationOutput struct {
	ID               string
	UserID           string
	IsAllEnabled     bool
	IsYoyoEnabled    bool
	IsReportEnabled  bool
	IsMessageEnabled bool
	Subscription     json.RawMessage
}

func NewCreateNotification(repo domain.NotificationRepository) *CreateNotification {
	return &CreateNotification{repo: repo}
}

func (u *CreateNotification) Execute(input NotificationInput) (domain.Notification, error) {
	if !isValidNotificationInput(input) {
		return domain.Notification{}, domain.ErrValidation
	}

	notification := domain.NewNotification(
		domain.NewNotificationID(),
		input.UserID,
		input.IsAllEnabled,
		input.IsYoyoEnabled,
		input.IsReportEnabled,
		input.IsMessageEnabled,
		input.Subscription,
	)

	return u.repo.Create(notification)
}

func isValidNotificationInput(input NotificationInput) bool {
	if !domain.IsValidUserID(input.UserID) {
		return false
	}
	if len(input.Subscription) == 0 || string(input.Subscription) == "null" {
		return false
	}
	return json.Valid(input.Subscription)
}
