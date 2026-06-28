package fincode

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pasokatazip/backend/internal/domain"
)

func TestCreateCustomer(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/customers" {
			t.Errorf("path = %s, want /v1/customers", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer secret" {
			t.Errorf("Authorization = %q", got)
		}
		if got := r.Header.Get("Api-Version"); got != defaultAPIVersion {
			t.Errorf("Api-Version = %q", got)
		}
		if got := r.Header.Get("idempotent_key"); got != "request-id" {
			t.Errorf("idempotent_key = %q", got)
		}

		var body createCustomerRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if body.ID != "user-id" || body.Email != "user@example.com" {
			t.Errorf("request body = %+v", body)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(createCustomerResponse{
			ID:    "user-id",
			Email: "user@example.com",
		})
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, SecretKey: "secret"})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	customer, err := client.CreateCustomer(context.Background(), domain.FincodeCustomerInput{
		ID:             "user-id",
		Email:          "user@example.com",
		IdempotencyKey: "request-id",
	})
	if err != nil {
		t.Fatalf("CreateCustomer: %v", err)
	}
	if customer.ID != "user-id" || customer.Email != "user@example.com" {
		t.Errorf("customer = %+v", customer)
	}
}

func TestCreateCustomerReturnsAPIError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"errors":[{"error_code":"E000000000"}]}`, http.StatusBadRequest)
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, SecretKey: "secret"})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = client.CreateCustomer(context.Background(), domain.FincodeCustomerInput{
		Email: "user@example.com",
	})
	var apiError *APIError
	if !errors.As(err, &apiError) {
		t.Fatalf("error = %v, want APIError", err)
	}
	if apiError.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", apiError.StatusCode)
	}
}
