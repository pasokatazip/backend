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

type GetPurchaseUsecase interface {
	Execute(domain.UserID) (onetime.FincodePurchaseStatus, error)
}

type PurchaseController struct {
	start StartPurchaseUsecase
	get   GetPurchaseUsecase
}

func NewPurchaseController(start StartPurchaseUsecase, get GetPurchaseUsecase) *PurchaseController {
	return &PurchaseController{start: start, get: get}
}

// Start starts a one-time purchase checkout.
// @Summary 買い切り決済開始
// @Description 認証中のユーザー向けにfincodeのカード登録URLを発行します。登録完了後に一回だけ請求します。
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

// Get returns the permanent purchase entitlement.
// @Summary 買い切り購入状態取得
// @Description 認証中のユーザーの買い切り購入状態を取得します。
// @Tags purchases
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.FincodePurchaseStatusResponse
// @Failure 401 {string} string "認証が必要"
// @Router /purchases [get]
func (c *PurchaseController) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(w, r)
	if !ok {
		return
	}
	status, err := c.get.Execute(userID)
	if err != nil {
		writeDomainError(w, err, "failed to get purchase")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dto.FincodePurchaseStatusResponse{
		Purchased:  status.Purchased,
		CustomerID: status.CustomerID,
		PaymentID:  status.PaymentID,
	})
}
