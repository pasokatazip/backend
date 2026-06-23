package usecases

import "github.com/pasokatazip/backend/internal/domain"

type FindNotificationByUserID struct {
	repo domain.NotificationRepository
}

func NewFindNotificationByUserID(repo domain.NotificationRepository) *FindNotificationByUserID {
	return &FindNotificationByUserID{repo: repo}
}

func (u *FindNotificationByUserID) Execute(userID domain.UserID) (domain.Notification, error) {
	if !domain.IsValidUserID(userID) {
		return domain.Notification{}, domain.ErrValidation
	}

	return u.repo.FindByUserID(userID)
}
