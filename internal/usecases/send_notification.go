package usecases

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pasokatazip/backend/internal/domain"
)

type SendNotification struct {
	repo   domain.NotificationRepository
	sender NotificationSender
}

type NotificationSender interface {
	Send(ctx context.Context, subscription json.RawMessage, payload NotificationPayload) error
}

type SendNotificationInput struct {
	Type  domain.NotificationType
	Title string
	Body  string
	Data  json.RawMessage
}

type NotificationPayload struct {
	Title string          `json:"title"`
	Body  string          `json:"body"`
	Data  json.RawMessage `json:"data,omitempty"`
}

type SendNotificationOutput struct {
	TargetCount int
	SentCount   int
	FailedCount int
	Errors      []string
}

func NewSendNotification(repo domain.NotificationRepository, sender NotificationSender) *SendNotification {
	return &SendNotification{repo: repo, sender: sender}
}

func (u *SendNotification) Execute(ctx context.Context, input SendNotificationInput) (SendNotificationOutput, error) {
	if !isValidSendNotificationInput(input) || u.sender == nil {
		return SendNotificationOutput{}, domain.ErrValidation
	}

	notifications, err := u.repo.FindEnabledForSend(input.Type)
	if err != nil {
		return SendNotificationOutput{}, err
	}

	output := SendNotificationOutput{TargetCount: len(notifications)}
	payload := NotificationPayload{
		Title: input.Title,
		Body:  input.Body,
		Data:  input.Data,
	}

	for _, notification := range notifications {
		if err := u.sender.Send(ctx, notification.Subscription(), payload); err != nil {
			output.FailedCount++
			output.Errors = append(output.Errors, fmt.Sprintf("%s: %v", notification.UserID(), err))
			continue
		}

		output.SentCount++
	}

	return output, nil
}

func isValidSendNotificationInput(input SendNotificationInput) bool {
	if input.Title == "" || input.Body == "" {
		return false
	}
	if len(input.Data) > 0 && !json.Valid(input.Data) {
		return false
	}

	switch input.Type {
	case domain.NotificationTypeYoyo, domain.NotificationTypeReport, domain.NotificationTypeMessage:
		return true
	default:
		return false
	}
}
