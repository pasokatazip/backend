package persistence

import (
	"database/sql"
	"errors"

	"github.com/pasokatazip/backend/internal/domain"
)

type NotificationRepository struct {
	DB *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{DB: db}
}

func (r *NotificationRepository) Create(notification domain.Notification) (domain.Notification, error) {
	query := `
		INSERT INTO notifications (
			id,
			user_id,
			is_all_enabled,
			is_yoyo_enabled,
			is_report_enabled,
			is_message_enabled,
			subscription
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.DB.Exec(
		query,
		notification.ID(),
		notification.UserID(),
		notification.IsAllEnabled(),
		notification.IsYoyoEnabled(),
		notification.IsReportEnabled(),
		notification.IsMessageEnabled(),
		notification.Subscription(),
	)
	if err != nil {
		return domain.Notification{}, err
	}

	return notification, nil
}

func (r *NotificationRepository) Update(notification domain.Notification) (domain.Notification, error) {
	query := `
		UPDATE notifications
		SET
			is_all_enabled = $1,
			is_yoyo_enabled = $2,
			is_report_enabled = $3,
			is_message_enabled = $4,
			subscription = $5
		WHERE id = $6
			AND user_id = $7
	`

	result, err := r.DB.Exec(
		query,
		notification.IsAllEnabled(),
		notification.IsYoyoEnabled(),
		notification.IsReportEnabled(),
		notification.IsMessageEnabled(),
		notification.Subscription(),
		notification.ID(),
		notification.UserID(),
	)
	if err != nil {
		return domain.Notification{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return domain.Notification{}, err
	}
	if rowsAffected == 0 {
		return domain.Notification{}, domain.ErrNotFound
	}

	return notification, nil
}

func (r *NotificationRepository) FindByUserID(userID domain.UserID) (domain.Notification, error) {
	query := `
		SELECT
			id,
			user_id,
			is_all_enabled,
			is_yoyo_enabled,
			is_report_enabled,
			is_message_enabled,
			subscription
		FROM notifications
		WHERE user_id = $1
	`

	return r.scanNotification(r.DB.QueryRow(query, userID))
}

func (r *NotificationRepository) scanNotification(row *sql.Row) (domain.Notification, error) {
	var (
		id               string
		userID           string
		isAllEnabled     bool
		isYoyoEnabled    bool
		isReportEnabled  bool
		isMessageEnabled bool
		subscription     string
	)

	if err := row.Scan(
		&id,
		&userID,
		&isAllEnabled,
		&isYoyoEnabled,
		&isReportEnabled,
		&isMessageEnabled,
		&subscription,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Notification{}, domain.ErrNotFound
		}
		return domain.Notification{}, err
	}

	return domain.NewNotification(
		domain.NotificationID(id),
		domain.UserID(userID),
		isAllEnabled,
		isYoyoEnabled,
		isReportEnabled,
		isMessageEnabled,
		subscription,
	), nil
}
