package presenter

import (
	"encoding/json"
	"testing"

	"github.com/pasokatazip/backend/internal/usecases"
)

func TestResponsePresentersDefineJSONContract(t *testing.T) {
	tests := []struct {
		name     string
		response any
		want     string
	}{
		{"posts", NewFindByPetPresenter().Output(nil), "null"},
		{"pet departure", NewUpdatePetDepartureStatusPresenter().Output(usecases.UpdatePetDepartureStatusOutput{}), `{"pet_id":"","status":"","eligible_at":"0001-01-01T00:00:00Z","scheduled_departure_at":"0001-01-01T00:00:00Z"}`},
		{"active evolution history", NewActivePetEvolutionHistoryPresenter().Output(usecases.FindActivePetEvolutionHistoryOutput{}), `{"pet_id":"","created_at":"0001-01-01T00:00:00Z","current_stage_key":"","stages":null,"evolutions":null}`},
		{"current evolution status", NewCurrentPetEvolutionStatusPresenter().Output(usecases.FindCurrentPetEvolutionStatusOutput{}), `{"pet_id":"","current_stage":{"id":0,"stage_key":"","stage_no":0,"name":""},"can_evolve":false,"next_stages":null}`},
		{"pet growth record", NewPetGrowthRecordPresenter().Output(usecases.FindPetGrowthRecordOutput{}), `{"pet_id":"","color":"","created_at":"0001-01-01T00:00:00Z","current_stage_id":0,"stages":null,"total_experience":0,"feed_count":0,"experience_events":null,"evolutions":null}`},
		{"historical reports", NewHistoricalReportPresenter().Output(nil), "null"},
		{"hourly simulation", NewRunHourlySimulationPresenter().Output(usecases.RunHourlyPetSimulationOutput{}), `{"simulated_at":"0001-01-01T00:00:00Z","total_pets":0,"processed":0,"skipped":0,"interest_propagations":0,"reports_created":0,"results":null}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			responseJSON, err := json.Marshal(tt.response)
			if err != nil {
				t.Fatalf("marshal response: %v", err)
			}

			if string(responseJSON) != tt.want {
				t.Fatalf("response JSON = %s, want %s", responseJSON, tt.want)
			}
		})
	}
}
