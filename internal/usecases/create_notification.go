package usecases

import "github.com/pasokatazip/backend/internal/domain"

type CreateNotification struct {
	repo domain.NotificationRepository
}

type NotificationInput struct {
	UserID           domain.UserID
	IsAllEnabled     bool
	IsYoyoEnabled    bool
	IsReportEnabled  bool
	IsMessageEnabled bool
	Subscription     string
}

type NotificationOutput struct {
	ID               string
	UserID           string
	IsAllEnabled     bool
	IsYoyoEnabled    bool
	IsReportEnabled  bool
	IsMessageEnabled bool
	Subscription     string
}

func NewCreateNotification(repo domain.NotificationRepository) *CreateNotification {
	return &CreateNotification{repo: repo}
}

func (u *CreateNotification) Execute(input NotificationInput) (domain.Notification, error) {
	if !domain.IsValidUserID(input.UserID) || input.Subscription == "" {
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
