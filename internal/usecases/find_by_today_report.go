package usecases

import (
	"time"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/timeutil"
)

type FindByDateReportInput struct {
	PetID      domain.PetID
	ReportDate *time.Time // nil の場合は前日（JST）を取得する。
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

type FindByDateReport struct {
	repo domain.ReportRepository
}

func NewFindByDate(repo domain.ReportRepository) *FindByDateReport {
	return &FindByDateReport{repo: repo}
}

// Execute は、指定日またはデフォルトの前日分（JST）のレポートを返す。
func (r *FindByDateReport) Execute(input FindByDateReportInput) ([]ReportOutput, error) {

	if input.PetID == "" || !domain.IsValidPetID(input.PetID) {
		return nil, domain.ErrValidation
	}

	reportDate := defaultReportDate(timeutil.NowJST())
	if input.ReportDate != nil {
		reportDate = input.ReportDate.In(timeutil.LocationJST())
	}

	reports, err := r.repo.FindByDate(input.PetID, reportDate)
	if err != nil {
		return nil, err
	}

	outputs := make([]ReportOutput, 0, len(reports))
	for _, report := range reports {
		outputs = append(outputs, reportOutput(report))
	}

	return outputs, nil
}

func defaultReportDate(now time.Time) time.Time {
	return now.In(timeutil.LocationJST()).AddDate(0, 0, -1)
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
