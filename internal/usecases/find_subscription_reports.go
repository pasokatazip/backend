package usecases

import (
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type FindSubscriptionReportsInput struct {
	UserID domain.UserID
	Date   time.Time
}

type SubscriptionReportsOutput struct {
	Reports []ReportOutput
	Pet     PetOutput
}

type SubscriptionReportRepository interface {
	FindByUserAndDate(domain.UserID, time.Time) ([]domain.Report, error)
}

type SubscriptionReportPetRepository interface {
	FindByID(domain.PetID) (domain.Pet, error)
}

type FindSubscriptionReports struct {
	reportRepo SubscriptionReportRepository
	petRepo    SubscriptionReportPetRepository
}

func NewFindSubscriptionReports(
	reportRepo SubscriptionReportRepository,
	petRepo SubscriptionReportPetRepository,
) *FindSubscriptionReports {
	return &FindSubscriptionReports{reportRepo: reportRepo, petRepo: petRepo}
}

func (u *FindSubscriptionReports) Execute(input FindSubscriptionReportsInput) (SubscriptionReportsOutput, error) {
	if !domain.IsValidUserID(input.UserID) || input.Date.IsZero() {
		return SubscriptionReportsOutput{}, domain.ErrValidation
	}

	reports, err := u.reportRepo.FindByUserAndDate(input.UserID, input.Date)
	if err != nil {
		return SubscriptionReportsOutput{}, err
	}
	if len(reports) == 0 {
		return SubscriptionReportsOutput{}, domain.ErrNotFound
	}

	petID := reports[0].PetID()
	for _, report := range reports[1:] {
		if report.PetID() != petID {
			return SubscriptionReportsOutput{}, domain.ErrInternal
		}
	}

	pet, err := u.petRepo.FindByID(petID)
	if err != nil {
		return SubscriptionReportsOutput{}, err
	}
	if pet.UserID() != input.UserID {
		return SubscriptionReportsOutput{}, domain.ErrUnauthorized
	}

	reportOutputs := make([]ReportOutput, 0, len(reports))
	for _, report := range reports {
		reportOutputs = append(reportOutputs, reportOutput(report))
	}

	return SubscriptionReportsOutput{
		Reports: reportOutputs,
		Pet: PetOutput{
			ID: string(pet.ID()), Name: pet.Name(), Color: pet.Color(), IsDeleted: pet.IsDeleted(),
			UserID: string(pet.UserID()), Energy: pet.Energy(), Curiosity: pet.Curiosity(),
			Sociality: pet.Sociality(), Routine: pet.Routine(),
			CurrentGroupMasterID: pet.CurrentGroupMasterID(), CurrentStageID: pet.CurrentStageID(),
			CreatedAt: pet.CreatedAt(), UpdatedAt: pet.UpdatedAt(),
		},
	}, nil
}
