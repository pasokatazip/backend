package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/usecases"
)

type ReportController struct {
	findByToday *usecases.FindByTodayReport
}

func NewReportController(findByToday *usecases.FindByTodayReport) *ReportController {
	return &ReportController{findByToday: findByToday}
}

// FindByToday ペットの当日レポートを取得します。
// @Summary 当日レポート取得
// @Description 指定したペットIDの当日分レポートを取得します。
// @Tags reports
// @Produce json
// @Param pet_id path string true "ペットID"
// @Success 200 {array} usecases.FindByTodayReportOutput "取得成功"
// @Failure 400 {string} string "ペットID不正"
// @Failure 500 {string} string "サーバーエラー"
// @Router /reports/{pet_id} [get]
func (c *ReportController) FindByToday(w http.ResponseWriter, r *http.Request) {
	petID := r.PathValue("pet_id")

	if petID == "" {
		http.Error(w, "missing pet_id", http.StatusBadRequest)
		return
	}

	outputs, err := c.findByToday.Execute(usecases.FindByTodayReportInput{PetID: domain.PetID(petID)})
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "failed to fetch reports", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(outputs)
}
