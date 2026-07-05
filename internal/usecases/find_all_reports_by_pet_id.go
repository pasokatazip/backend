package usecases

import "github.com/pasokatazip/backend/internal/domain"

type FindAllReportsByPetIDInput struct {
	PetID domain.PetID
}

type FindAllReportsByPetID struct {
	repo domain.ReportRepository
}

func NewFindAllReportsByPetID(repo domain.ReportRepository) *FindAllReportsByPetID {
	return &FindAllReportsByPetID{repo: repo}
}

func (u *FindAllReportsByPetID) Execute(input FindAllReportsByPetIDInput) ([]FindByTodayReportOutput, error) {
	if !domain.IsValidPetID(input.PetID) {
		return nil, domain.ErrValidation
	}

	reports, err := u.repo.FindAllByPetID(input.PetID)
	if err != nil {
		return nil, err
	}

	outputs := make([]FindByTodayReportOutput, 0, len(reports))
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
