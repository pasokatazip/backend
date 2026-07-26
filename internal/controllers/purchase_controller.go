package controllers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pasokatazip/backend/internal/controllers/dto"
	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/usecases/onetime"
)

type StartPurchaseUsecase interface {
	Execute(context.Context, domain.UserID) (domain.FincodeCardSession, error)
}

type ConfirmPurchaseUsecase interface {
	Execute(context.Context, domain.UserID) (onetime.FincodePurchaseStatus, error)
}

type PurchaseController struct {
	start   StartPurchaseUsecase
	confirm ConfirmPurchaseUsecase
}

func NewPurchaseController(
	start StartPurchaseUsecase,
	confirm ConfirmPurchaseUsecase,
) *PurchaseController {
	return &PurchaseController{start: start, confirm: confirm}
}

// Start starts a one-time purchase checkout.
// @Summary 買い切りカード登録開始
// @Description 認証中のユーザー向けにfincodeのカード登録URLを発行します。
// @Tags purchases
// @Produce json
// @Security BearerAuth
// @Success 201 {object} dto.FincodeCheckoutResponse
// @Failure 400 {string} string "リクエスト不正"
// @Failure 401 {string} string "認証が必要"
// @Failure 409 {string} string "購入済み"
// @Failure 502 {string} string "外部サービスエラー"
// @Router /purchases/checkout [post]
func (c *PurchaseController) Start(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(w, r)
	if !ok {
		return
	}
	session, err := c.start.Execute(r.Context(), userID)
	if err != nil {
		writeDomainError(w, err, "failed to start purchase")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(dto.NewFincodeCheckoutResponse(session))
}

// Confirm charges the registered card and grants the permanent entitlement.
// @Summary 買い切り決済確定
// @Description fincodeで登録済みのカードに一括決済を実行し、購入権限を有効化します。
// @Tags purchases
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.FincodePurchaseConfirmResponse
// @Failure 400 {string} string "リクエスト不正"
// @Failure 401 {string} string "認証が必要"
// @Failure 404 {string} string "登録カードが見つかりません"
// @Failure 502 {string} string "決済サービスエラー"
// @Router /purchases/confirm [post]
func (c *PurchaseController) Confirm(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(w, r)
	if !ok {
		return
	}
	status, err := c.confirm.Execute(r.Context(), userID)
	if err != nil {
		writeDomainError(w, err, "failed to confirm purchase")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dto.FincodePurchaseConfirmResponse{
		Subsc: status.Purchased,
	})
}
