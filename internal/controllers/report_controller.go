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

type ReportController struct {
	findByDate       *usecases.FindByDateReport
	findAllByPetID   *usecases.FindAllReportsByPetID
	findSubscription *usecases.FindSubscriptionReports
}

func NewReportController(
	findByDate *usecases.FindByDateReport,
	findAllByPetID *usecases.FindAllReportsByPetID,
	findSubscription *usecases.FindSubscriptionReports,
) *ReportController {
	return &ReportController{
		findByDate:       findByDate,
		findAllByPetID:   findAllByPetID,
		findSubscription: findSubscription,
	}
}

type SubscriptionReportRequest struct {
	Date string `json:"date"`
}

// FindSubscription returns reports for the date and pet contained in the JWT.
// @Summary 契約ユーザーの日別レポート取得
// @Tags reports
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SubscriptionReportRequest true "取得日 (YYYY-MM-DD)"
// @Success 200 {object} dto.SubscriptionReportsResponse
// @Router /suubsc/report [post]
func (c *ReportController) FindSubscription(w http.ResponseWriter, r *http.Request) {
	userID, userOK := middleware.GetUserID(r.Context())
	petID, petOK := middleware.GetPetID(r.Context())
	if !userOK || !petOK {
		http.Error(w, domain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	var request SubscriptionReportRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "body must contain date in YYYY-MM-DD format", http.StatusBadRequest)
		return
	}
	reportDate, err := time.ParseInLocation("2006-01-02", request.Date, timeutil.LocationJST())
	if err != nil {
		http.Error(w, "date must be YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	output, err := c.findSubscription.Execute(usecases.FindSubscriptionReportsInput{
		UserID: domain.UserID(userID),
		PetID:  domain.PetID(petID),
		Date:   reportDate,
	})
	if err != nil {
		writeDomainError(w, err, "failed to fetch reports")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewSubscriptionReportsResponse(output))
}

// FindAllByPetID returns all reports for a pet in reverse chronological order.
// @Summary ペットの全レポート取得
// @Description 指定したペットIDに紐づく全レポートを日時の新しい順で取得します。
// @Tags reports
// @Produce json
// @Security BearerAuth
// @Param pet_id path string true "ペットID"
// @Success 200 {array} usecases.FindByTodayReportOutput "取得成功"
// @Failure 400 {string} string "ペットID不正"
// @Failure 401 {string} string "認証が必要"
// @Failure 403 {string} string "サブスクリプションが必要"
// @Failure 500 {string} string "サーバーエラー"
// @Router /subsc/reports/{pet_id} [get]
func (c *ReportController) FindAllByPetID(w http.ResponseWriter, r *http.Request) {
	petID := domain.PetID(r.PathValue("pet_id"))
	outputs, err := c.findAllByPetID.Execute(usecases.FindAllReportsByPetIDInput{PetID: petID})
	if err != nil {
		writeDomainError(w, err, "failed to fetch reports")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(outputs)
}

// FindByDate ペットの指定日レポートを取得します。date 未指定時は前日（JST）を返します。
// @Summary 日次レポート取得
// @Description 指定したペットIDの指定日分レポートを取得します。date を省略した場合は前日（JST）分を返します。
// @Tags reports
// @Produce json
// @Param pet_id path string true "ペットID"
// @Param date query string false "取得日（YYYY-MM-DD、未指定時は前日・JST）"
// @Success 200 {object} dto.ReportsResponse "取得成功"
// @Failure 400 {string} string "ペットID不正"
// @Failure 500 {string} string "サーバーエラー"
// @Router /reports/{pet_id} [get]
func (c *ReportController) FindByDate(w http.ResponseWriter, r *http.Request) {
	petID := r.PathValue("pet_id")

	if petID == "" {
		http.Error(w, "missing pet_id", http.StatusBadRequest)
		return
	}

	var reportDate *time.Time
	if rawDate := r.URL.Query().Get("date"); rawDate != "" {
		parsedDate, err := time.ParseInLocation("2006-01-02", rawDate, timeutil.LocationJST())
		if err != nil {
			http.Error(w, "date must be YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		reportDate = &parsedDate
	}

	outputs, err := c.findByDate.Execute(usecases.FindByDateReportInput{
		PetID:      domain.PetID(petID),
		ReportDate: reportDate,
	})
	if err != nil {
		writeDomainError(w, err, "failed to fetch reports")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewReportsResponse(outputs))
}
