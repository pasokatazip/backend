package controllers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pasokatazip/backend/internal/controllers/dto"
	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/infrastructure/middleware"
	"github.com/pasokatazip/backend/internal/usecases"
)

type StartSubscriptionUsecase interface {
	Execute(ctx context.Context, userID domain.UserID) (domain.FincodeCardSession, error)
}

type CancelSubscriptionUsecase interface {
	Execute(ctx context.Context, userID domain.UserID) error
}

type GetSubscriptionUsecase interface {
	Execute(userID domain.UserID) (usecases.FincodeSubscriptionStatus, error)
}

type SubscriptionController struct {
	start  StartSubscriptionUsecase
	cancel CancelSubscriptionUsecase
	get    GetSubscriptionUsecase
}

func NewSubscriptionController(
	start StartSubscriptionUsecase,
	cancel CancelSubscriptionUsecase,
	get GetSubscriptionUsecase,
) *SubscriptionController {
	return &SubscriptionController{start: start, cancel: cancel, get: get}
}

// Get サブスクリプション状態を取得します。
// @Summary サブスクリプション状態取得
// @Description 認証中のユーザーのサブスクリプション契約状態を取得します。
// @Tags subscriptions
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.FincodeSubscriptionStatusResponse "取得成功"
// @Failure 400 {string} string "リクエスト不正"
// @Failure 401 {string} string "認証が必要"
// @Failure 404 {string} string "契約情報が見つからない"
// @Failure 409 {string} string "契約状態が競合"
// @Failure 502 {string} string "外部サービスエラー"
// @Router /subscriptions [get]
func (c *SubscriptionController) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(w, r)
	if !ok {
		return
	}

	status, err := c.get.Execute(userID)
	if err != nil {
		writeSubscriptionError(w, err, "failed to get subscription")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewFincodeSubscriptionStatusResponse(status))
}

// Start サブスクリプションの決済を開始します。
// @Summary サブスクリプション決済開始
// @Description 認証中のユーザー向けに fincode の決済ページを作成します。
// @Tags subscriptions
// @Produce json
// @Security BearerAuth
// @Success 201 {object} dto.FincodeCheckoutResponse "決済ページ作成成功"
// @Failure 400 {string} string "リクエスト不正"
// @Failure 401 {string} string "認証が必要"
// @Failure 404 {string} string "ユーザーが見つからない"
// @Failure 409 {string} string "すでに契約済み"
// @Failure 502 {string} string "外部サービスエラー"
// @Router /subscriptions/checkout [post]
func (c *SubscriptionController) Start(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(w, r)
	if !ok {
		return
	}

	session, err := c.start.Execute(r.Context(), userID)
	if err != nil {
		writeSubscriptionError(w, err, "failed to start subscription")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.NewFincodeCheckoutResponse(session))
}

// Cancel サブスクリプションを解約します。
// @Summary サブスクリプション解約
// @Description 認証中のユーザーのサブスクリプションを解約します。
// @Tags subscriptions
// @Security BearerAuth
// @Success 202 "解約リクエスト受理"
// @Failure 400 {string} string "リクエスト不正"
// @Failure 401 {string} string "認証が必要"
// @Failure 404 {string} string "契約情報が見つからない"
// @Failure 409 {string} string "契約状態が競合"
// @Failure 502 {string} string "外部サービスエラー"
// @Router /subscriptions [delete]
func (c *SubscriptionController) Cancel(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(w, r)
	if !ok {
		return
	}

	if err := c.cancel.Execute(r.Context(), userID); err != nil {
		writeSubscriptionError(w, err, "failed to cancel subscription")
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func authenticatedUserID(w http.ResponseWriter, r *http.Request) (domain.UserID, bool) {
	value, ok := middleware.GetUserID(r.Context())
	if !ok || !domain.IsValidUserID(domain.UserID(value)) {
		http.Error(w, domain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return "", false
	}
	return domain.UserID(value), true
}

func writeSubscriptionError(w http.ResponseWriter, err error, fallback string) {
	writeDomainError(w, err, fallback)
}
