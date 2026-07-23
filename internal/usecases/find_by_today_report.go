package usecases

import (
	"github.com/pasokatazip/backend/internal/domain"
	"time"
)

type FindByTodayReportInput struct {
	PetID domain.PetID
}

type ReportOutput struct {
	ID        string
	PetID     string
	GroupName string
	CreatedAt time.Time
	Gossip    string
	HourSlot  int
	Souvenirs []SouvenirOutput
	Rumors    []string
}

type FindByTodayReportOutput struct {
	ID        string
	PetID     domain.PetID
	HourSlot  int
	Gossip    string
	GroupName string `json:"Group_name"`
	CreatedAt time.Time
	Rumors    []string `json:"rumors"`
}

type SouvenirOutput struct {
	ID          string
	DisplayName string
	ImageURL    string
}

type FindByTodayReport struct {
	repo domain.ReportRepository
}

func NewFindByToDay(repo domain.ReportRepository) *FindByTodayReport {
	return &FindByTodayReport{repo: repo}
}

func (r *FindByTodayReport) Execute(input FindByTodayReportInput) ([]ReportOutput, error) {

	if input.PetID == "" || !domain.IsValidPetID(input.PetID) {
		return nil, domain.ErrValidation
	}

	reports, err := r.repo.FindByToday(input.PetID)
	if err != nil {
		return nil, err
	}

	outputs := make([]ReportOutput, 0, len(reports))
	for _, report := range reports {
		outputs = append(outputs, reportOutput(report))
	}

	return outputs, nil
}

func reportOutput(report domain.Report) ReportOutput {
	souvenirs := make([]SouvenirOutput, 0, len(report.Souvenirs()))
	for _, souvenir := range report.Souvenirs() {
		souvenirs = append(souvenirs, SouvenirOutput{
			ID: souvenir.ID(), DisplayName: souvenir.DisplayName(), ImageURL: souvenir.ImageURL(),
		})
	}
	return ReportOutput{
		ID: string(report.ID()), PetID: string(report.PetID()), GroupName: report.GroupName(),
		CreatedAt: report.CreatedAt(), Gossip: report.Gossip(), HourSlot: report.HourSlot(),
		Souvenirs: souvenirs,
		Rumors:    report.Rumors(),
	}
}
