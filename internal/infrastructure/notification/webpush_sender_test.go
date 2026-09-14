package notification

import (
	"encoding/json"
	"testing"

	"github.com/pasokatazip/backend/internal/usecases"
)

func TestEncodeWebPushPayloadUsesExternalJSONContract(t *testing.T) {
	body, err := encodeWebPushPayload(usecases.NotificationPayload{
		Title: "タイトル",
		Body:  "本文",
		Data:  json.RawMessage(`{"type":"report"}`),
	})
	if err != nil {
		t.Fatalf("encodeWebPushPayload() error = %v", err)
	}

	want := `{"title":"タイトル","body":"本文","data":{"type":"report"}}`
	if string(body) != want {
		t.Fatalf("encodeWebPushPayload() = %s, want %s", body, want)
	}
}
