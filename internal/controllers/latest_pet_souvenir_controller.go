package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/pasokatazip/backend/internal/controllers/dto"
	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/infrastructure/middleware"
	"github.com/pasokatazip/backend/internal/usecases"
)

type LatestPetSouvenirController struct {
	findLatest *usecases.FindLatestPetSouvenir
}

func NewLatestPetSouvenirController(
	findLatest *usecases.FindLatestPetSouvenir,
) *LatestPetSouvenirController {
	return &LatestPetSouvenirController{findLatest: findLatest}
}

// Find returns the latest souvenir found by the authenticated user's active pet.
// @Summary アクティブペットの最新おみやげ取得
// @Description 認証中ユーザーのアクティブペットが最後に見つけたおみやげを取得します。未取得の場合はsouvenirがnullになります。
// @Tags pets
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.LatestPetSouvenirResponse "取得成功"
// @Failure 400 {string} string "ユーザーID不正"
// @Failure 401 {string} string "認証が必要"
// @Failure 404 {string} string "アクティブペットが見つからない"
// @Failure 500 {string} string "サーバーエラー"
// @Router /pets/me/souvenirs/latest [get]
func (c *LatestPetSouvenirController) Find(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, domain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	output, err := c.findLatest.Execute(usecases.FindLatestPetSouvenirInput{
		UserID: domain.UserID(userID),
	})
	if err != nil {
		writeDomainError(w, err, "failed to fetch latest pet souvenir")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewLatestPetSouvenirResponse(output))
}
