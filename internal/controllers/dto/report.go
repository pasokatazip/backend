package dto

import (
	"time"

	"github.com/pasokatazip/backend/internal/usecases"
)

type ReportsResponse struct {
	Reports    []ReportResponse `json:"reports"`
	HasPraised bool             `json:"hasPraised"`
}

type SubscriptionReportsResponse struct {
	Reports    []ReportResponse      `json:"reports"`
	Pet        SubscriptionReportPet `json:"pet"`
	HasPraised bool                  `json:"hasPraised"`
}

type SubscriptionReportPet struct {
	ID              string    `json:"pet_id"`
	Name            string    `json:"name"`
	Color           string    `json:"color"`
	CurrentStageKey string    `json:"current_stage_key"`
	CurrentStageNo  int       `json:"current_stage_no"`
	IsDeleted       bool      `json:"is_deleted"`
	CreatedAt       time.Time `json:"created_at"`
}

type ReportResponse struct {
	ID        string             `json:"id"`
	PetID     string             `json:"petID"`
	GroupName string             `json:"groupName"`
	CreatedAt time.Time          `json:"createdAt"`
	Gossip    string             `json:"gossip"`
	HourSlot  int                `json:"hourSlot"`
	Souvenirs []SouvenirResponse `json:"souvenirs"`
	Rumors    []string           `json:"rumors"`
}

type SouvenirResponse struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	ImageURL    string `json:"imageURL"`
}

func NewReportsResponse(output usecases.FindByDateReportOutput) ReportsResponse {
	return ReportsResponse{
		Reports:    newReportResponses(output.Reports),
		HasPraised: output.HasPraised,
	}
}

func newReportResponses(outputs []usecases.ReportOutput) []ReportResponse {
	reports := make([]ReportResponse, 0, len(outputs))
	for _, output := range outputs {
		souvenirs := make([]SouvenirResponse, 0, len(output.Souvenirs))
		for _, souvenir := range output.Souvenirs {
			souvenirs = append(souvenirs, SouvenirResponse{
				ID: souvenir.ID, DisplayName: souvenir.DisplayName, ImageURL: souvenir.ImageURL,
			})
		}
		reports = append(reports, ReportResponse{
			ID: output.ID, PetID: output.PetID, GroupName: output.GroupName,
			CreatedAt: output.CreatedAt, Gossip: output.Gossip, HourSlot: output.HourSlot,
			Souvenirs: souvenirs,
			Rumors:    output.Rumors,
		})
	}
	return reports
}

func NewSubscriptionReportsResponse(output usecases.SubscriptionReportsOutput) SubscriptionReportsResponse {
	return SubscriptionReportsResponse{
		Reports:    newReportResponses(output.Reports),
		HasPraised: output.HasPraised,
		Pet: SubscriptionReportPet{
			ID: output.Pet.ID, Name: output.Pet.Name, Color: output.Pet.Color,
			CurrentStageKey: output.Pet.CurrentStageKey,
			CurrentStageNo:  output.Pet.CurrentStageNo,
			IsDeleted:       output.Pet.IsDeleted,
			CreatedAt:       output.Pet.CreatedAt,
		},
	}
}
