package presenter

import (
	"github.com/pasokatazip/backend/internal/controllers/dto"
	"github.com/pasokatazip/backend/internal/usecases"
)

type ReportPresenter struct{}

func NewReportPresenter() *ReportPresenter {
	return &ReportPresenter{}
}

func (p *ReportPresenter) Daily(output usecases.FindByDateReportOutput) dto.ReportsResponse {
	return dto.ReportsResponse{
		Reports:    newReportResponses(output.Reports),
		HasPraised: output.HasPraised,
	}
}

func (p *ReportPresenter) Subscription(output usecases.SubscriptionReportsOutput) dto.SubscriptionReportsResponse {
	return dto.SubscriptionReportsResponse{
		Reports:    newReportResponses(output.Reports),
		HasPraised: output.HasPraised,
		Pet: dto.SubscriptionReportPet{
			ID: output.Pet.ID, Name: output.Pet.Name, Color: output.Pet.Color,
			CurrentStageKey: output.Pet.CurrentStageKey,
			CurrentStageNo:  output.Pet.CurrentStageNo,
			IsDeleted:       output.Pet.IsDeleted,
			CreatedAt:       output.Pet.CreatedAt,
		},
	}
}

func newReportResponses(outputs []usecases.ReportOutput) []dto.ReportResponse {
	reports := make([]dto.ReportResponse, 0, len(outputs))
	for _, output := range outputs {
		souvenirs := make([]dto.SouvenirResponse, 0, len(output.Souvenirs))
		for _, souvenir := range output.Souvenirs {
			souvenirs = append(souvenirs, dto.SouvenirResponse{
				ID: souvenir.ID, DisplayName: souvenir.DisplayName, ImageURL: souvenir.ImageURL,
			})
		}
		reports = append(reports, dto.ReportResponse{
			ID: output.ID, PetID: output.PetID, GroupName: output.GroupName,
			CreatedAt: output.CreatedAt, Gossip: output.Gossip, HourSlot: output.HourSlot,
			Souvenirs: souvenirs,
			Rumors:    output.Rumors,
		})
	}
	return reports
}
