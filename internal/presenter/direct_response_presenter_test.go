package presenter

import (
	"encoding/json"
	"testing"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/usecases"
)

func TestResponsePresentersPreserveJSONContract(t *testing.T) {
	tests := []struct {
		name     string
		original any
		response any
	}{
		{"posts", []domain.Post(nil), NewFindByPetPresenter().Output(nil)},
		{"pet departure", usecases.UpdatePetDepartureStatusOutput{}, NewUpdatePetDepartureStatusPresenter().Output(usecases.UpdatePetDepartureStatusOutput{})},
		{"active evolution history", usecases.FindActivePetEvolutionHistoryOutput{}, NewActivePetEvolutionHistoryPresenter().Output(usecases.FindActivePetEvolutionHistoryOutput{})},
		{"current evolution status", usecases.FindCurrentPetEvolutionStatusOutput{}, NewCurrentPetEvolutionStatusPresenter().Output(usecases.FindCurrentPetEvolutionStatusOutput{})},
		{"pet growth record", usecases.FindPetGrowthRecordOutput{}, NewPetGrowthRecordPresenter().Output(usecases.FindPetGrowthRecordOutput{})},
		{"historical reports", []usecases.FindByTodayReportOutput(nil), NewHistoricalReportPresenter().Output(nil)},
		{"hourly simulation", usecases.RunHourlyPetSimulationOutput{}, NewRunHourlySimulationPresenter().Output(usecases.RunHourlyPetSimulationOutput{})},
	}

	for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
			originalJSON, err := json.Marshal(tt.original)
			if err != nil {
				t.Fatalf("marshal original: %v", err)
			}

			responseJSON, err := json.Marshal(tt.response)
			if err != nil {
				t.Fatalf("marshal response: %v", err)
			}

			if string(responseJSON) != string(originalJSON) {
				t.Fatalf("response JSON = %s, want %s", responseJSON, originalJSON)
			}
		})
	}
}
