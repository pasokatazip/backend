package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/pasokatazip/backend/internal/controllers/dto"
	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/infrastructure/middleware"
	"github.com/pasokatazip/backend/internal/timeutil"
	"github.com/pasokatazip/backend/internal/usecases"
)

type SouvenirPraiseFlagController struct {
	mark *usecases.MarkSouvenirPraised
}

func NewSouvenirPraiseFlagController(
	mark *usecases.MarkSouvenirPraised,
) *SouvenirPraiseFlagController {
	return &SouvenirPraiseFlagController{mark: mark}
}

// Mark records the authenticated user's first praise selection for a report date.
// @Summary おみやげの「ほめる」選択済み記録
// @Description 二回以降も成功し、初回の選択時刻を保持します。
// @Tags reports
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.SouvenirPraiseFlagResponse
// @Failure 401 {string} string "認証が必要"
// @Failure 500 {string} string "サーバーエラー"
// @Param date path string true "レポート対象日。YYYY-MM-DD形式"
// @Failure 400 {string} string "日付またはユーザーID不正"
// @Router /users/me/souvenir-praise/{date} [put]
func (c *SouvenirPraiseFlagController) Mark(w http.ResponseWriter, r *http.Request) {
	userID, ok := souvenirPraiseUserID(w, r)
	if !ok {
		return
	}

	reportDate, err := time.ParseInLocation(
		"2006-01-02",
		r.PathValue("date"),
		timeutil.LocationJST(),
	)
	if err != nil {
		http.Error(w, "date must be YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	output, err := c.mark.Execute(usecases.MarkSouvenirPraisedInput{
		UserID:     userID,
		ReportDate: reportDate,
	})
	if err != nil {
		writeDomainError(w, err, "failed to mark souvenir praise flag")
		return
	}
	c.writeResponse(w, output)
}

func souvenirPraiseUserID(w http.ResponseWriter, r *http.Request) (domain.UserID, bool) {
	value, ok := middleware.GetUserID(r.Context())
	if !ok || !domain.IsValidUserID(domain.UserID(value)) {
		http.Error(w, domain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return "", false
	}
	return domain.UserID(value), true
}

func (c *SouvenirPraiseFlagController) writeResponse(
	w http.ResponseWriter,
	output usecases.SouvenirPraiseFlagOutput,
) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(dto.NewSouvenirPraiseFlagResponse(output)); err != nil {
		return
	}
}
