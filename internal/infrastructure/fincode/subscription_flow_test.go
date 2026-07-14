package fincode

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

func TestCreateCardSession(t *testing.T) {
	t.Parallel()

	location := time.FixedZone("Asia/Tokyo", 9*60*60)
	expiresAt := time.Date(2026, 6, 27, 18, 30, 0, 0, location)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/card_sessions" || r.Method != http.MethodPost {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		var request createCardSessionRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if request.CustomerID != "customer-id" || request.Expire != "2026/06/27 18:30:00" {
			t.Errorf("request = %+v", request)
		}
		if request.GuideMailSendFlag != "0" || request.CompletionMailSendFlag != "0" {
			t.Errorf("mail flags = %q, %q", request.GuideMailSendFlag, request.CompletionMailSendFlag)
		}

		json.NewEncoder(w).Encode(createCardSessionResponse{
			ID:      "lk_test",
			LinkURL: "https://secure.test.fincode.jp/v1/links/lk_test",
			Expire:  request.Expire,
		})
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, SecretKey: "secret"})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	session, err := client.CreateCardSession(context.Background(), domain.FincodeCardSessionInput{
		CustomerID:      "customer-id",
		ShopServiceName: "Example",
		ExpiresAt:       expiresAt,
	})
	if err != nil {
		t.Fatalf("CreateCardSession: %v", err)
	}
	if session.ID != "lk_test" || session.LinkURL == "" || !session.ExpiresAt.Equal(expiresAt) {
		t.Errorf("session = %+v", session)
	}
}

func TestCreateAndCancelSubscription(t *testing.T) {
	t.Parallel()

	var canceled bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/subscriptions":
			if got := r.Header.Get("idempotent_key"); got != "request-id" {
				t.Errorf("idempotent_key = %q", got)
			}
			var request createSubscriptionRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatalf("decode request: %v", err)
			}
			if request.PayType != "Card" || request.PlanID != "plan-id" || request.CardID != "card-id" || request.StartDate != "2026/06/27" {
				t.Errorf("request = %+v", request)
			}
			json.NewEncoder(w).Encode(createSubscriptionResponse{
				ID:         "subscription-id",
				PlanID:     request.PlanID,
				CustomerID: request.CustomerID,
				Status:     "ACTIVE",
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/v1/subscriptions/subscription-id":
			if r.URL.Query().Get("pay_type") != "Card" {
				t.Errorf("pay_type = %q", r.URL.Query().Get("pay_type"))
			}
			canceled = true
			json.NewEncoder(w).Encode(map[string]string{"id": "subscription-id"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, SecretKey: "secret"})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	subscription, err := client.CreateSubscription(context.Background(), domain.FincodeSubscriptionInput{
		PlanID:         "plan-id",
		CustomerID:     "customer-id",
		CardID:         "card-id",
		StartDate:      time.Date(2026, 6, 27, 0, 0, 0, 0, time.UTC),
		IdempotencyKey: "request-id",
	})
	if err != nil {
		t.Fatalf("CreateSubscription: %v", err)
	}
	if subscription.ID != "subscription-id" || subscription.Status != "ACTIVE" {
		t.Errorf("subscription = %+v", subscription)
	}
	if err := client.CancelSubscription(context.Background(), subscription.ID); err != nil {
		t.Fatalf("CancelSubscription: %v", err)
	}
	if !canceled {
		t.Error("cancel API was not called")
	}
}
