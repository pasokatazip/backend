package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/pasokatazip/backend/internal/usecases"
)

type FincodeController struct {
	handleCardRegist         HandleCardRegistUsecase
	handleSubscriptionRegist HandleSubscriptionRegistUsecase
	handleSubscriptionCancel HandleSubscriptionCancelUsecase
}

type HandleCardRegistUsecase interface {
	Execute(input usecases.CardRegistrationInput) error
}

type HandleSubscriptionRegistUsecase interface {
	Execute(input usecases.SubscRegistrationInput) error
}

type HandleSubscriptionCancelUsecase interface {
	Execute(input usecases.SubscCancelInput) error
}

func NewWebhookController(
	handleCardRegist HandleCardRegistUsecase,
	handleSubscriptionRegist HandleSubscriptionRegistUsecase,
	handleSubscriptionCancel HandleSubscriptionCancelUsecase,
) *FincodeController {
	return &FincodeController{
		handleCardRegist:         handleCardRegist,
		handleSubscriptionRegist: handleSubscriptionRegist,
		handleSubscriptionCancel: handleSubscriptionCancel,
	}
}

type WebhookEvent struct {
	Event          string `json:"event"`
	CustomerID     string `json:"customer_id"`
	CardID         string `json:"card_id"`
	SubscriptionID string `json:"subscription_id"`
	Status         string `json:"status"`
}

func (c *FincodeController) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event WebhookEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "invalid webhook payload", http.StatusBadRequest)
		return
	}

	var err error

	switch event.Event {
	case "card.regist":
		err = c.handleCardRegist.Execute(usecases.CardRegistrationInput{
			CustomerID: event.CustomerID,
		})

	case "subscription.card.regist", "subscription.card.update":
		err = c.handleSubscriptionRegist.Execute(usecases.SubscRegistrationInput{
			CustomerID:     event.CustomerID,
			SubscriptionID: event.SubscriptionID,
			Status:         event.Status,
		})

	case "subscription.card.delete", "subscription.card.cancel":
		err = c.handleSubscriptionCancel.Execute(usecases.SubscCancelInput{
			CustomerID:     event.CustomerID,
			SubscriptionID: event.SubscriptionID,
		})

	//webHookの再送を加味して200でコントローラー側は200でreturn
	default:
		w.WriteHeader(http.StatusOK)
		return
	}

	if err != nil {
		http.Error(w, "failed to handle webhook", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
