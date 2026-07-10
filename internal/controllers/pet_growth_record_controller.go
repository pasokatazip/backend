package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/usecases"
)

type PetGrowthRecordController struct {
	findByPetID *usecases.FindPetGrowthRecord
}

func NewPetGrowthRecordController(findByPetID *usecases.FindPetGrowthRecord) *PetGrowthRecordController {
	return &PetGrowthRecordController{findByPetID: findByPetID}
}

// FindByPetID returns a pet's growth record.
// @Summary ペットの成長記録取得
// @Description 指定したペットIDに紐づく経験値集計、経験値取得イベント、進化履歴を取得します。
// @Tags pets
// @Produce json
// @Security BearerAuth
// @Param pet_id path string true "ペットID"
// @Success 200 {object} usecases.FindPetGrowthRecordOutput "取得成功"
// @Failure 400 {string} string "ペットID不正"
// @Failure 401 {string} string "認証が必要"
// @Failure 403 {string} string "サブスクリプションが必要"
// @Failure 500 {string} string "サーバーエラー"
// @Router /subsc/growth_records/{pet_id} [get]
func (c *PetGrowthRecordController) FindByPetID(w http.ResponseWriter, r *http.Request) {
	petID := domain.PetID(r.PathValue("pet_id"))

	output, err := c.findByPetID.Execute(usecases.FindPetGrowthRecordInput{PetID: petID})
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "failed to fetch pet growth record", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}
