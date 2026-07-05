package usecases

import (
	"github.com/pasokatazip/backend/internal/domain"
	"time"
)

type FindByTodayReportInput struct {
	PetID domain.PetID
}

type FindByTodayReportOutput struct {
	ID        string
	PetID     domain.PetID
	HourSlot  int
	Gossip    string
	GroupName string `json:"Group_name"`
	CreatedAt time.Time
}

type FindByTodayReport struct {
	repo domain.ReportRepository
}

func NewFindByToDay(repo domain.ReportRepository) *FindByTodayReport {
	return &FindByTodayReport{repo: repo}
}

func (r *FindByTodayReport) Execute(input FindByTodayReportInput) ([]FindByTodayReportOutput, error) {

	if input.PetID == "" || !domain.IsValidPetID(input.PetID) {
		return nil, domain.ErrValidation
	}

	reports, err := r.repo.FindByToday(input.PetID)
	if err != nil {
		return nil, err
	}

	var outputs []FindByTodayReportOutput
	for _, report := range reports {
		outputs = append(outputs, FindByTodayReportOutput{
			ID:        string(report.ID()),
			PetID:     report.PetID(),
			HourSlot:  report.HourSlot(),
			Gossip:    report.Gossip(),
			GroupName: report.GroupName(),
			CreatedAt: report.CreatedAt(),
		})
	}

	return outputs, nil
}
