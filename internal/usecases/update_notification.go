package usecases

import (
	"encoding/json"

	"github.com/pasokatazip/backend/internal/domain"
)

type UpdateNotification struct {
	repo domain.NotificationRepository
}

type UpdateNotificationInput struct {
	UserID           domain.UserID
	IsAllEnabled     bool
	IsYoyoEnabled    bool
	IsReportEnabled  bool
	IsMessageEnabled bool
	Subscription     json.RawMessage
}

func NewUpdateNotification(repo domain.NotificationRepository) *UpdateNotification {
	return &UpdateNotification{repo: repo}
}

func (u *UpdateNotification) Execute(input UpdateNotificationInput) (domain.Notification, error) {
	if !isValidUpdateNotificationInput(input) {
		return domain.Notification{}, domain.ErrValidation
	}

	current, err := u.repo.FindByUserID(input.UserID)
	if err != nil {
		return domain.Notification{}, err
	}

	notification := domain.NewNotification(
		current.ID(),
		input.UserID,
		input.IsAllEnabled,
		input.IsYoyoEnabled,
		input.IsReportEnabled,
		input.IsMessageEnabled,
		notificationSubscription(input.Subscription, current.Subscription()),
	)

	return u.repo.Update(notification)
}

func isValidUpdateNotificationInput(input UpdateNotificationInput) bool {
	if !domain.IsValidUserID(input.UserID) {
		return false
	}
	if len(input.Subscription) == 0 {
		return true
	}
	if string(input.Subscription) == "null" {
		return false
	}
	return json.Valid(input.Subscription)
}

func notificationSubscription(input json.RawMessage, current json.RawMessage) json.RawMessage {
	if len(input) == 0 {
		return current
	}
	return input
}
