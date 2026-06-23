package presenter

import (
	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/usecases"
)

type NotificationPresenter struct{}

func NewNotificationPresenter() *NotificationPresenter {
	return &NotificationPresenter{}
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
