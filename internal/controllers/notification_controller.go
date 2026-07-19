package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/pasokatazip/backend/internal/controllers/dto"
	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/infrastructure/middleware"
	"github.com/pasokatazip/backend/internal/presenter"
	"github.com/pasokatazip/backend/internal/usecases"
)

type NotificationController struct {
	createNotification       *usecases.CreateNotification
	updateNotification       *usecases.UpdateNotification
	findNotificationByUserID *usecases.FindNotificationByUserID
}

func NewNotificationController(
	createNotification *usecases.CreateNotification,
	updateNotification *usecases.UpdateNotification,
	findNotificationByUserID *usecases.FindNotificationByUserID,
) *NotificationController {
	return &NotificationController{
		createNotification:       createNotification,
		updateNotification:       updateNotification,
		findNotificationByUserID: findNotificationByUserID,
	}
}

// Create 通知設定を新規登録します。
// @Summary 通知設定登録
// @Description 認証中のユーザーの通知設定と Web Push 購読情報を登録します。
// @Tags notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.NotificationRequest true "通知設定"
// @Success 201 {object} dto.NotificationResponse "登録成功"
// @Failure 400 {string} string "リクエスト不正"
// @Failure 401 {string} string "認証が必要"
// @Failure 404 {string} string "対象が見つからない"
// @Failure 500 {string} string "サーバーエラー"
// @Router /notifications [post]
func (c *NotificationController) Create(w http.ResponseWriter, r *http.Request) {
	req, userID, ok := c.requestInput(w, r)
	if !ok {
		return
	}

	notification, err := c.createNotification.Execute(req.ToUseCaseInput(userID))
	if err != nil {
		c.handleError(w, err, "failed to create notification")
		return
	}

	c.writeNotification(w, http.StatusCreated, notification)
}

// Update 通知設定を更新します。
// @Summary 通知設定更新
// @Description 認証中のユーザーの通知設定と Web Push 購読情報を更新します。
// @Tags notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpdateNotificationRequest true "通知設定"
// @Success 200 {object} dto.NotificationResponse "更新成功"
// @Failure 400 {string} string "リクエスト不正"
// @Failure 401 {string} string "認証が必要"
// @Failure 404 {string} string "通知設定が見つからない"
// @Failure 500 {string} string "サーバーエラー"
// @Router /notifications [put]
func (c *NotificationController) Update(w http.ResponseWriter, r *http.Request) {
	req, userID, ok := c.updateRequestInput(w, r)
	if !ok {
		return
	}

	notification, err := c.updateNotification.Execute(req.ToUseCaseInput(userID))
	if err != nil {
		c.handleError(w, err, "failed to update notification")
		return
	}

	c.writeNotification(w, http.StatusOK, notification)
}

// FindByUserID 通知設定を取得します。
// @Summary 通知設定取得
// @Description 認証中のユーザーの通知設定を取得します。
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.NotificationResponse "取得成功"
// @Failure 401 {string} string "認証が必要"
// @Failure 404 {string} string "通知設定が見つからない"
// @Failure 500 {string} string "サーバーエラー"
// @Router /notifications [get]
func (c *NotificationController) FindByUserID(w http.ResponseWriter, r *http.Request) {
	userIDString, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, domain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	notification, err := c.findNotificationByUserID.Execute(domain.UserID(userIDString))
	if err != nil {
		c.handleError(w, err, "failed to fetch notification")
		return
	}

	c.writeNotification(w, http.StatusOK, notification)
}

func (c *NotificationController) requestInput(w http.ResponseWriter, r *http.Request) (dto.NotificationRequest, domain.UserID, bool) {
	var req dto.NotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return dto.NotificationRequest{}, "", false
	}

	userIDString, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, domain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return dto.NotificationRequest{}, "", false
	}

	return req, domain.UserID(userIDString), true
}

func (c *NotificationController) updateRequestInput(w http.ResponseWriter, r *http.Request) (dto.UpdateNotificationRequest, domain.UserID, bool) {
	var req dto.UpdateNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return dto.UpdateNotificationRequest{}, "", false
	}

	userIDString, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, domain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return dto.UpdateNotificationRequest{}, "", false
	}

	return req, domain.UserID(userIDString), true
}

func (c *NotificationController) writeNotification(w http.ResponseWriter, status int, notification domain.Notification) {
	pr := presenter.NewNotificationPresenter()
	output := pr.Output(notification)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(dto.NewNotificationResponse(output))
}

func (c *NotificationController) handleError(w http.ResponseWriter, err error, fallback string) {
	writeDomainError(w, err, fallback)
}
