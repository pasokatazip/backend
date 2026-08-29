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

type FindByDateReportOutput struct {
	Reports    []ReportOutput
	HasPraised bool
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
	reportRepo domain.ReportRepository
	praiseRepo domain.SouvenirPraiseFlagRepository
}

func NewFindByDate(
	reportRepo domain.ReportRepository,
	praiseRepo domain.SouvenirPraiseFlagRepository,
) *FindByDateReport {
	return &FindByDateReport{reportRepo: reportRepo, praiseRepo: praiseRepo}
}

// Execute は、指定日またはデフォルトの前日分（JST）のレポートを返す。
func (r *FindByDateReport) Execute(input FindByDateReportInput) (FindByDateReportOutput, error) {

	if input.PetID == "" || !domain.IsValidPetID(input.PetID) {
		return FindByDateReportOutput{}, domain.ErrValidation
	}

	reportDate := defaultReportDate(timeutil.NowJST())
	if input.ReportDate != nil {
		reportDate = input.ReportDate.In(timeutil.LocationJST())
	}

	reports, err := r.reportRepo.FindByDate(input.PetID, reportDate)
	if err != nil {
		return FindByDateReportOutput{}, err
	}

	praiseFlag, err := r.praiseRepo.FindByPetIDAndDate(input.PetID, reportDate)
	if err != nil {
		return FindByDateReportOutput{}, err
	}

	outputs := make([]ReportOutput, 0, len(reports))
	for _, report := range reports {
		outputs = append(outputs, reportOutput(report))
	}

	return FindByDateReportOutput{
		Reports:    outputs,
		HasPraised: praiseFlag.HasPraised(),
	}, nil
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
