package usecases

import "github.com/pasokatazip/backend/internal/domain"

type UpdateNotification struct {
	repo domain.NotificationRepository
}

func NewUpdateNotification(repo domain.NotificationRepository) *UpdateNotification {
	return &UpdateNotification{repo: repo}
}

func (u *UpdateNotification) Execute(input NotificationInput) (domain.Notification, error) {
	if !domain.IsValidUserID(input.UserID) || input.Subscription == "" {
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
		input.Subscription,
	)

	return u.repo.Update(notification)
}
