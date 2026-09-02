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
	findLatest           *usecases.FindLatestPetSouvenir
	findLatestHistorical *usecases.FindLatestHistoricalPetSouvenir
}

func NewLatestPetSouvenirController(
	findLatest *usecases.FindLatestPetSouvenir,
	findLatestHistorical *usecases.FindLatestHistoricalPetSouvenir,
) *LatestPetSouvenirController {
	return &LatestPetSouvenirController{
		findLatest:           findLatest,
		findLatestHistorical: findLatestHistorical,
	}
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

// FindHistorical returns the latest souvenir found by one of the authenticated
// user's historical pets.
// @Summary 過去ペットの最後のおみやげ取得
// @Description 認証中ユーザーが所有する過去ペットが最後に見つけたおみやげを取得します。未取得の場合はsouvenirがnullになります。
// @Tags pets
// @Produce json
// @Security BearerAuth
// @Param pet_id path string true "過去ペットID"
// @Success 200 {object} dto.LatestPetSouvenirResponse "取得成功"
// @Failure 400 {string} string "ユーザーIDまたはペットID不正"
// @Failure 401 {string} string "認証が必要"
// @Failure 403 {string} string "サブスクリプションが必要"
// @Failure 404 {string} string "過去ペットが見つからない"
// @Failure 500 {string} string "サーバーエラー"
// @Router /subsc/pets/{pet_id}/souvenirs/latest [get]
func (c *LatestPetSouvenirController) FindHistorical(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, domain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	output, err := c.findLatestHistorical.Execute(usecases.FindLatestHistoricalPetSouvenirInput{
		UserID: domain.UserID(userID),
		PetID:  domain.PetID(r.PathValue("pet_id")),
	})
	if err != nil {
		writeDomainError(w, err, "failed to fetch latest historical pet souvenir")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewLatestPetSouvenirResponse(output))
}
