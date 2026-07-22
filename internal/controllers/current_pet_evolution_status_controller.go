package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/infrastructure/middleware"
	"github.com/pasokatazip/backend/internal/usecases"
)

type CurrentPetEvolutionStatusController struct {
	findStatus *usecases.FindCurrentPetEvolutionStatus
}

func NewCurrentPetEvolutionStatusController(
	findStatus *usecases.FindCurrentPetEvolutionStatus,
) *CurrentPetEvolutionStatusController {
	return &CurrentPetEvolutionStatusController{findStatus: findStatus}
}

// Find returns the active pet's evolution readiness without committing one.
// @Summary アクティブペットの進化可能状態を取得
// @Description 進化条件と進化先候補を確認します。このAPI自体は進化を実行しません。
// @Tags pets
// @Produce json
// @Security BearerAuth
// @Success 200 {object} usecases.FindCurrentPetEvolutionStatusOutput "取得成功"
// @Failure 400 {string} string "ユーザーID不正"
// @Failure 401 {string} string "認証が必要"
// @Failure 500 {string} string "サーバーエラー"
// @Router /pets/evolution-status [get]
func (c *CurrentPetEvolutionStatusController) Find(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, domain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	output, err := c.findStatus.Execute(usecases.FindCurrentPetEvolutionStatusInput{
		UserID: domain.UserID(userID),
	})
	if err != nil {
		writeDomainError(w, err, "failed to fetch current pet evolution status")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}
