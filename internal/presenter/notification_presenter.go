package presenter

import (
	"github.com/pasokatazip/backend/internal/controllers/dto"
	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/usecases"
)

type NotificationPresenter struct{}

func NewNotificationPresenter() *NotificationPresenter {
	return &NotificationPresenter{}
}

func (p *NotificationPresenter) Response(notification domain.Notification) dto.NotificationResponse {
	output := p.Output(notification)
	return dto.NotificationResponse{
		ID:               output.ID,
		UserID:           output.UserID,
		IsAllEnabled:     output.IsAllEnabled,
		IsYoyoEnabled:    output.IsYoyoEnabled,
		IsReportEnabled:  output.IsReportEnabled,
		IsMessageEnabled: output.IsMessageEnabled,
		Subscription:     output.Subscription,
	}
}

func (p *NotificationPresenter) Output(notification domain.Notification) usecases.NotificationOutput {
	return usecases.NotificationOutput{
		ID:               string(notification.ID()),
		UserID:           string(notification.UserID()),
		IsAllEnabled:     notification.IsAllEnabled(),
		IsYoyoEnabled:    notification.IsYoyoEnabled(),
		IsReportEnabled:  notification.IsReportEnabled(),
		IsMessageEnabled: notification.IsMessageEnabled(),
		Subscription:     notification.Subscription(),
	}
}
