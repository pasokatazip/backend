package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pasokatazip/backend/internal/usecases"
	"github.com/pasokatazip/backend/internal/usecases/subsc"
)

type cardRegistWebhookStub struct {
	calls int
	input usecases.CardRegistrationInput
}

func (s *cardRegistWebhookStub) Execute(_ context.Context, input usecases.CardRegistrationInput) error {
	s.calls++
	s.input = input
	return nil
}

type subscriptionRegistWebhookStub struct{}

func (subscriptionRegistWebhookStub) Execute(subsc.SubscRegistrationInput) error { return nil }

type subscriptionCancelWebhookStub struct{}

func (subscriptionCancelWebhookStub) Execute(subsc.SubscCancelInput) error { return nil }

func TestFincodeWebhookHandlesActivatedRedirectCard(t *testing.T) {
	cardRegist := &cardRegistWebhookStub{}
	controller := NewWebhookController(
		cardRegist,
		subscriptionRegistWebhookStub{},
		subscriptionCancelWebhookStub{},
		"webhook-signature",
	)
	body := `{
		"event":"customers.payment_methods.updated",
		"customer_id":"customer-id",
		"card_id":"card-id",
		"card_status":"ACTIVATED",
		"pay_type":"Card"
	}`
	req := httptest.NewRequest(http.MethodPost, "/webhooks/fincode", strings.NewReader(body))
	req.Header.Set("Fincode-Signature", "webhook-signature")
	res := httptest.NewRecorder()

	controller.Handle(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
	}
	if cardRegist.calls != 1 || cardRegist.input.CustomerID != "customer-id" || cardRegist.input.CardID != "card-id" {
		t.Fatalf("card registration = calls:%d input:%+v", cardRegist.calls, cardRegist.input)
	}
	assertWebhookSuccess(t, res)
}

func TestFincodeWebhookIgnoresCardBeforeActivation(t *testing.T) {
	cardRegist := &cardRegistWebhookStub{}
	controller := NewWebhookController(
		cardRegist,
		subscriptionRegistWebhookStub{},
		subscriptionCancelWebhookStub{},
		"webhook-signature",
	)
	body := `{
		"event":"customers.payment_methods.updated",
		"customer_id":"customer-id",
		"card_id":"card-id",
		"card_status":"AWAITING_CUSTOMER_ACTION",
		"pay_type":"Card"
	}`
	req := httptest.NewRequest(http.MethodPost, "/webhooks/fincode", strings.NewReader(body))
	req.Header.Set("Fincode-Signature", "webhook-signature")
	res := httptest.NewRecorder()

	controller.Handle(res, req)

	if cardRegist.calls != 0 {
		t.Fatalf("card registration calls = %d, want 0", cardRegist.calls)
	}
	assertWebhookSuccess(t, res)
}

func TestFincodeWebhookReturnsSuccessBodyForUnknownEvent(t *testing.T) {
	controller := NewWebhookController(
		&cardRegistWebhookStub{},
		subscriptionRegistWebhookStub{},
		subscriptionCancelWebhookStub{},
		"webhook-signature",
	)
	req := httptest.NewRequest(http.MethodPost, "/webhooks/fincode", strings.NewReader(`{"event":"unknown"}`))
	req.Header.Set("Fincode-Signature", "webhook-signature")
	res := httptest.NewRecorder()

	controller.Handle(res, req)

	assertWebhookSuccess(t, res)
}

func assertWebhookSuccess(t *testing.T, res *httptest.ResponseRecorder) {
	t.Helper()
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
	}
	var response map[string]string
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response["receive"] != "0" {
		t.Fatalf("response = %+v, want receive 0", response)
	}
}
