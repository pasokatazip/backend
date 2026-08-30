package dto

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/pasokatazip/backend/internal/usecases"
)

func TestNewReportsResponseJSONShape(t *testing.T) {
	response := NewReportsResponse(usecases.FindByDateReportOutput{
		HasPraised: true,
		Reports: []usecases.ReportOutput{{
			ID: "report-id", PetID: "pet-id", GroupName: "公園の群れ", CreatedAt: time.Now(),
			Gossip: "gossip", HourSlot: 12, Souvenirs: []usecases.SouvenirOutput{},
			Rumors: []string{"近くでゲームの話をしていた"},
		}},
	})
	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	for _, field := range []string{`"reports"`, `"hasPraised":true`, `"petID"`, `"groupName"`, `"createdAt"`, `"hourSlot"`, `"souvenirs":[]`, `"rumors":["近くでゲームの話をしていた"]`} {
		if !strings.Contains(string(encoded), field) {
			t.Fatalf("JSON = %s, missing %s", encoded, field)
		}
	}
}

func TestNewReportsResponseUsesEmptyReportsArray(t *testing.T) {
	encoded, err := json.Marshal(NewReportsResponse(usecases.FindByDateReportOutput{}))
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if string(encoded) != `{"reports":[],"hasPraised":false}` {
		t.Fatalf("JSON = %s, want reports array", encoded)
	}
}
