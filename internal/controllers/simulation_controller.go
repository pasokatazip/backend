package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/timeutil"
	"github.com/pasokatazip/backend/internal/usecases"
)

type SimulationController struct {
	runHourly *usecases.RunHourlyPetSimulation
}

type runHourlySimulationRequest struct {
	SimulatedAt *string `json:"simulated_at"`
}

func NewSimulationController(runHourly *usecases.RunHourlyPetSimulation) *SimulationController {
	return &SimulationController{runHourly: runHourly}
}

func (c *SimulationController) RunHourly(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req runHourlySimulationRequest
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
	}

	var simulatedAt *time.Time
	if req.SimulatedAt != nil && *req.SimulatedAt != "" {
		parsed, err := time.Parse(time.RFC3339, *req.SimulatedAt)
		if err != nil {
			http.Error(w, "simulated_at must be RFC3339", http.StatusBadRequest)
			return
		}
		parsed = parsed.In(timeutil.LocationJST())
		simulatedAt = &parsed
	}

	output, err := c.runHourly.Execute(usecases.RunHourlyPetSimulationInput{
		SimulatedAt: simulatedAt,
	})
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			http.Error(w, "active group_masters are required", http.StatusBadRequest)
			return
		}
		http.Error(w, "failed to run hourly simulation", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}
