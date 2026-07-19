package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/infrastructure/middleware"
	"github.com/pasokatazip/backend/internal/usecases"
)

type ActivePetEvolutionHistoryController struct {
	findByActivePet *usecases.FindActivePetEvolutionHistory
}

func NewActivePetEvolutionHistoryController(
	findByActivePet *usecases.FindActivePetEvolutionHistory,
) *ActivePetEvolutionHistoryController {
	return &ActivePetEvolutionHistoryController{findByActivePet: findByActivePet}
}

// Find returns evolution history for the authenticated user's active pet.
// @Summary アクティブペットの進化履歴取得
// @Description 認証中ユーザーの現在アクティブなペットに紐づく進化履歴を取得します。
// @Tags pets
// @Produce json
// @Security BearerAuth
// @Success 200 {object} usecases.FindActivePetEvolutionHistoryOutput "取得成功"
// @Failure 400 {string} string "ユーザーID不正"
// @Failure 401 {string} string "認証が必要"
// @Failure 500 {string} string "サーバーエラー"
// @Router /pets/evolutions [get]
func (c *ActivePetEvolutionHistoryController) Find(w http.ResponseWriter, r *http.Request) {
	userIDString, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, domain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	output, err := c.findByActivePet.Execute(usecases.FindActivePetEvolutionHistoryInput{
		UserID: domain.UserID(userIDString),
	})
	if err != nil {
		writeDomainError(w, err, "failed to fetch active pet evolution history")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}
