package controllers

import (
	"net/http"
	"errors"
	"encoding/json"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/usecases"
)

type ReportController struct {
	findByToday *usecases.FindByTodayReport
}

func NewReportController(findByToday *usecases.FindByTodayReport) *ReportController {
	return &ReportController{findByToday: findByToday}
}

func (c *ReportController)FindByToday( w http.ResponseWriter, r *http.Request ) {
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