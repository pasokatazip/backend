package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/infrastructure/middleware"
	"github.com/pasokatazip/backend/internal/usecases"
)

type PetGrowthRecordController struct {
	findByPetID *usecases.FindPetGrowthRecord
}

func NewPetGrowthRecordController(findByPetID *usecases.FindPetGrowthRecord) *PetGrowthRecordController {
	return &PetGrowthRecordController{findByPetID: findByPetID}
}

// FindByPetID returns a pet's evolution history and growth record.
// @Summary ペットの進化履歴取得
// @Description 指定したペットIDに紐づく段階一覧、進化履歴、経験値集計、経験値取得イベントを取得します。
// @Tags pets
// @Produce json
// @Security BearerAuth
// @Param pet_id path string true "ペットID"
// @Success 200 {object} usecases.FindPetGrowthRecordOutput "取得成功"
// @Failure 400 {string} string "ペットID不正"
// @Failure 401 {string} string "認証が必要"
// @Failure 403 {string} string "サブスクリプションが必要"
// @Failure 500 {string} string "サーバーエラー"
// @Router /subsc/pets/{pet_id}/evolutions [get]
func (c *PetGrowthRecordController) FindByPetID(w http.ResponseWriter, r *http.Request) {
	userIDString, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, domain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	petID := domain.PetID(r.PathValue("pet_id"))

	output, err := c.findByPetID.Execute(usecases.FindPetGrowthRecordInput{
		PetID:  petID,
		UserID: domain.UserID(userIDString),
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrValidation):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, domain.ErrUnauthorized):
			http.Error(w, domain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		default:
			http.Error(w, "failed to fetch pet growth record", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}
