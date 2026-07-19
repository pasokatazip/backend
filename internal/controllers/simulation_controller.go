package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/pasokatazip/backend/internal/timeutil"
	"github.com/pasokatazip/backend/internal/usecases"
)

type SimulationController struct {
	runHourly *usecases.RunHourlyPetSimulation
}

type RunHourlySimulationRequest struct {
	SimulatedAt *string `json:"simulated_at"`
}

func NewSimulationController(runHourly *usecases.RunHourlyPetSimulation) *SimulationController {
	return &SimulationController{runHourly: runHourly}
}

// RunHourly ペットの時間単位シミュレーションを実行します。
// @Summary 時間単位シミュレーション実行
// @Description 全アクティブペットを対象に、指定時刻または現在時刻の時間単位シミュレーションを実行します。
// @Tags simulations
// @Accept json
// @Produce json
// @Param request body RunHourlySimulationRequest false "シミュレーション条件"
// @Success 200 {object} usecases.RunHourlyPetSimulationOutput "実行成功"
// @Failure 400 {string} string "リクエストまたはマスターデータ不正"
// @Failure 405 {string} string "許可されていないメソッド"
// @Failure 500 {string} string "サーバーエラー"
// @Router /simulations/hourly [post]
func (c *SimulationController) RunHourly(w http.ResponseWriter, r *http.Request) {
	var req RunHourlySimulationRequest
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
		writeDomainError(w, err, "failed to run hourly simulation")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}
