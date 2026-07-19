package persistence

import (
	"database/sql"
	"encoding/json"
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
		) VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb)
	`

	_, err := r.DB.Exec(
		query,
		notification.ID(),
		notification.UserID(),
		notification.IsAllEnabled(),
		notification.IsYoyoEnabled(),
		notification.IsReportEnabled(),
		notification.IsMessageEnabled(),
		string(notification.Subscription()),
	)
	if err != nil {
		return domain.Notification{}, mapPersistenceError(err)
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
			subscription = $5::jsonb
		WHERE id = $6
			AND user_id = $7
	`

	result, err := r.DB.Exec(
		query,
		notification.IsAllEnabled(),
		notification.IsYoyoEnabled(),
		notification.IsReportEnabled(),
		notification.IsMessageEnabled(),
		string(notification.Subscription()),
		notification.ID(),
		notification.UserID(),
	)
	if err != nil {
		return domain.Notification{}, mapPersistenceError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return domain.Notification{}, mapPersistenceError(err)
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

func (r *NotificationRepository) FindEnabledForSend(notificationType domain.NotificationType) ([]domain.Notification, error) {
	enabledColumn, ok := notificationEnabledColumn(notificationType)
	if !ok {
		return nil, domain.ErrValidation
	}

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
		WHERE is_all_enabled = true
			AND ` + enabledColumn + ` = true
	`

	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, mapPersistenceError(err)
	}
	defer rows.Close()

	var notifications []domain.Notification
	for rows.Next() {
		notification, err := scanNotificationRows(rows)
		if err != nil {
			return nil, mapPersistenceError(err)
		}
		notifications = append(notifications, notification)
	}
	if err := rows.Err(); err != nil {
		return nil, mapPersistenceError(err)
	}

	return notifications, nil
}

func notificationEnabledColumn(notificationType domain.NotificationType) (string, bool) {
	switch notificationType {
	case domain.NotificationTypeYoyo:
		return "is_yoyo_enabled", true
	case domain.NotificationTypeReport:
		return "is_report_enabled", true
	case domain.NotificationTypeMessage:
		return "is_message_enabled", true
	default:
		return "", false
	}
}

func (r *NotificationRepository) scanNotification(row *sql.Row) (domain.Notification, error) {
	var (
		id               string
		userID           string
		isAllEnabled     bool
		isYoyoEnabled    bool
		isReportEnabled  bool
		isMessageEnabled bool
		subscription     []byte
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
		return domain.Notification{}, mapPersistenceError(err)
	}

	return domain.NewNotification(
		domain.NotificationID(id),
		domain.UserID(userID),
		isAllEnabled,
		isYoyoEnabled,
		isReportEnabled,
		isMessageEnabled,
		json.RawMessage(subscription),
	), nil
}

func scanNotificationRows(rows *sql.Rows) (domain.Notification, error) {
	var (
		id               string
		userID           string
		isAllEnabled     bool
		isYoyoEnabled    bool
		isReportEnabled  bool
		isMessageEnabled bool
		subscription     []byte
	)

	if err := rows.Scan(
		&id,
		&userID,
		&isAllEnabled,
		&isYoyoEnabled,
		&isReportEnabled,
		&isMessageEnabled,
		&subscription,
	); err != nil {
		return domain.Notification{}, mapPersistenceError(err)
	}

	return domain.NewNotification(
		domain.NotificationID(id),
		domain.UserID(userID),
		isAllEnabled,
		isYoyoEnabled,
		isReportEnabled,
		isMessageEnabled,
		json.RawMessage(subscription),
	), nil
}
