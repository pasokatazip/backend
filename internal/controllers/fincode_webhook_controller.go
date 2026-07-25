package controllers

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/pasokatazip/backend/internal/usecases"
	"github.com/pasokatazip/backend/internal/usecases/subsc"
)

type FincodeController struct {
	handleCardRegist         HandleCardRegistUsecase
	handleSubscriptionRegist HandleSubscriptionRegistUsecase
	handleSubscriptionCancel HandleSubscriptionCancelUsecase
	webhookSignature         string
}

type HandleCardRegistUsecase interface {
	Execute(ctx context.Context, input usecases.CardRegistrationInput) error
}

type HandleSubscriptionRegistUsecase interface {
	Execute(input subsc.SubscRegistrationInput) error
}

type HandleSubscriptionCancelUsecase interface {
	Execute(input subsc.SubscCancelInput) error
}

func NewWebhookController(
	handleCardRegist HandleCardRegistUsecase,
	handleSubscriptionRegist HandleSubscriptionRegistUsecase,
	handleSubscriptionCancel HandleSubscriptionCancelUsecase,
	webhookSignature string,
) *FincodeController {
	return &FincodeController{
		handleCardRegist:         handleCardRegist,
		handleSubscriptionRegist: handleSubscriptionRegist,
		handleSubscriptionCancel: handleSubscriptionCancel,
		webhookSignature:         webhookSignature,
	}
}

type WebhookEvent struct {
	Event          string `json:"event"`
	CustomerID     string `json:"customer_id"`
	CardID         string `json:"card_id"`
	SubscriptionID string `json:"subscription_id"`
	Status         string `json:"status"`
	CardStatus     string `json:"card_status"`
	PayType        string `json:"pay_type"`
}

// Handle fincode の Webhook を処理します。
// @Summary fincode Webhook受信
// @Description fincode から送信されるカード登録・契約登録・契約解約イベントを処理します。
// @Tags webhooks
// @Accept json
// @Produce json
// @Param Fincode-Signature header string true "Webhook署名"
// @Param request body WebhookEvent true "Webhookイベント"
// @Success 200 {object} map[string]string "受信成功"
// @Failure 400 {string} string "ペイロード不正"
// @Failure 401 {string} string "署名不正"
// @Failure 405 {string} string "許可されていないメソッド"
// @Failure 500 {string} string "処理失敗"
// @Router /webhooks/fincode [post]
func (c *FincodeController) Handle(w http.ResponseWriter, r *http.Request) {
	if c.webhookSignature == "" || subtle.ConstantTimeCompare(
		[]byte(r.Header.Get("Fincode-Signature")),
		[]byte(c.webhookSignature),
	) != 1 {
		http.Error(w, "invalid webhook signature", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var event WebhookEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "invalid webhook payload", http.StatusBadRequest)
		return
	}

	var err error

	switch event.Event {
	case "customers.payment_methods.created",
		"customers.payment_methods.updated",
		"customers.payment_methods.activated":
		// A hosted card-session can create, update, or activate a payment
		// method. The billing use case is idempotent across these variants.
		if !strings.EqualFold(event.PayType, "Card") || !strings.EqualFold(event.CardStatus, "ACTIVATED") {
			writeWebhookSuccess(w)
			return
		}
		err = c.handleCardRegist.Execute(r.Context(), usecases.CardRegistrationInput{
			CustomerID: event.CustomerID,
			CardID:     event.CardID,
		})

	case "card.regist":
		// Keep supporting the legacy card registration event for direct card API flows.
		err = c.handleCardRegist.Execute(r.Context(), usecases.CardRegistrationInput{
			CustomerID: event.CustomerID,
			CardID:     event.CardID,
		})

	case "subscription.card.regist", "subscription.card.update":
		if c.handleSubscriptionRegist == nil {
			writeWebhookSuccess(w)
			return
		}
		err = c.handleSubscriptionRegist.Execute(subsc.SubscRegistrationInput{
			CustomerID:     event.CustomerID,
			SubscriptionID: event.SubscriptionID,
			Status:         event.Status,
		})

	case "subscription.card.delete", "subscription.card.cancel":
		if c.handleSubscriptionCancel == nil {
			writeWebhookSuccess(w)
			return
		}
		err = c.handleSubscriptionCancel.Execute(subsc.SubscCancelInput{
			CustomerID:     event.CustomerID,
			SubscriptionID: event.SubscriptionID,
		})

	//webHookの再送を加味して200でコントローラー側は200でreturn
	default:
		writeWebhookSuccess(w)
		return
	}

	if err != nil {
		log.Printf("failed to handle fincode webhook event=%q: %v", event.Event, err)
		writeDomainError(w, err, "failed to handle webhook")
		return
	}

	writeWebhookSuccess(w)
}

func writeWebhookSuccess(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"receive": "0"})
}
