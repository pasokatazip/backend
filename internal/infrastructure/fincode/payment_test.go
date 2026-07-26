package fincode

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pasokatazip/backend/internal/domain"
)

func TestCreateAndExecutePayment(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/payments/order-id":
			_ = json.NewEncoder(w).Encode(paymentResponse{ID: "order-id", Status: "UNPROCESSED"})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/payments":
			var body createPaymentRequest
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body.ID != "order-id" || body.Amount != "1980" || body.JobCode != "CAPTURE" {
				t.Errorf("create body = %+v", body)
			}
			_ = json.NewEncoder(w).Encode(paymentResponse{ID: "order-id", Status: "UNPROCESSED"})
		case r.Method == http.MethodPut && r.URL.Path == "/v1/payments/order-id":
			var body executePaymentRequest
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body.CustomerID != "customer-id" || body.CardID != "card-id" {
				t.Errorf("execute body = %+v", body)
			}
			_ = json.NewEncoder(w).Encode(paymentResponse{ID: "order-id", Status: "CAPTURED"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, SecretKey: "secret"})
	if err != nil {
		t.Fatal(err)
	}
	found, err := client.GetPayment(context.Background(), "order-id")
	if err != nil || found.ID != "order-id" {
		t.Fatalf("GetPayment = %+v, %v", found, err)
	}
	created, err := client.CreatePayment(context.Background(), domain.FincodePaymentInput{
		ID: "order-id", Amount: 1980,
	})
	if err != nil || created.ID != "order-id" {
		t.Fatalf("CreatePayment = %+v, %v", created, err)
	}
	executed, err := client.ExecutePayment(context.Background(), domain.FincodePaymentInput{
		ID: "order-id", CustomerID: "customer-id", CardID: "card-id",
	})
	if err != nil || executed.Status != "CAPTURED" {
		t.Fatalf("ExecutePayment = %+v, %v", executed, err)
	}
}
