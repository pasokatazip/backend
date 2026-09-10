package usecases

import "github.com/pasokatazip/backend/internal/domain"

type FindAllReportsByPetIDInput struct {
	UserID domain.UserID
	PetID  domain.PetID
}

type FindAllReportsByPetID struct {
	reportRepo domain.ReportRepository
	petRepo    domain.PetRepository
}

func NewFindAllReportsByPetID(reportRepo domain.ReportRepository, petRepo domain.PetRepository) *FindAllReportsByPetID {
	return &FindAllReportsByPetID{reportRepo: reportRepo, petRepo: petRepo}
}

func (u *FindAllReportsByPetID) Execute(input FindAllReportsByPetIDInput) ([]FindByTodayReportOutput, error) {
	if _, err := findOwnedPet(u.petRepo, input.UserID, input.PetID); err != nil {
		return nil, err
	}

	reports, err := u.reportRepo.FindAllByPetID(input.PetID)
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
			Rumors:    report.Rumors(),
		})
	}

	return outputs, nil
}
