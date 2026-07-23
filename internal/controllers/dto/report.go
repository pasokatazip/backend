package dto

import (
	"time"

	"github.com/pasokatazip/backend/internal/usecases"
)

type ReportsResponse struct {
	Reports []ReportResponse `json:"reports"`
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

func NewReportsResponse(outputs []usecases.ReportOutput) ReportsResponse {
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
	return ReportsResponse{Reports: reports}
}
