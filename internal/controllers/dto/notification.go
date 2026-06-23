package dto

import (
	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/usecases"
)

type NotificationRequest struct {
	IsAllEnabled     bool   `json:"is_all_enabled"`
	IsYoyoEnabled    bool   `json:"is_yoyo_enabled"`
	IsReportEnabled  bool   `json:"is_report_enabled"`
	IsMessageEnabled bool   `json:"is_message_enabled"`
	Subscription     string `json:"subscription"`
}

func (r NotificationRequest) ToUseCaseInput(userID domain.UserID) usecases.NotificationInput {
	return usecases.NotificationInput{
		UserID:           userID,
		IsAllEnabled:     r.IsAllEnabled,
		IsYoyoEnabled:    r.IsYoyoEnabled,
		IsReportEnabled:  r.IsReportEnabled,
		IsMessageEnabled: r.IsMessageEnabled,
		Subscription:     r.Subscription,
	}
}

type NotificationResponse struct {
	ID               string `json:"id"`
	UserID           string `json:"user_id"`
	IsAllEnabled     bool   `json:"is_all_enabled"`
	IsYoyoEnabled    bool   `json:"is_yoyo_enabled"`
	IsReportEnabled  bool   `json:"is_report_enabled"`
	IsMessageEnabled bool   `json:"is_message_enabled"`
	Subscription     string `json:"subscription"`
}

func NewNotificationResponse(output usecases.NotificationOutput) NotificationResponse {
	return NotificationResponse{
		ID:               output.ID,
		UserID:           output.UserID,
		IsAllEnabled:     output.IsAllEnabled,
		IsYoyoEnabled:    output.IsYoyoEnabled,
		IsReportEnabled:  output.IsReportEnabled,
		IsMessageEnabled: output.IsMessageEnabled,
		Subscription:     output.Subscription,
	}
}
