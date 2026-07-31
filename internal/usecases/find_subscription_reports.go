package usecases

import (
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type FindSubscriptionReportsInput struct {
	UserID domain.UserID
	PetID  domain.PetID
	Date   time.Time
}

type SubscriptionReportsOutput struct {
	Reports []ReportOutput
	Pet     PetOutput
}

type SubscriptionReportRepository interface {
	FindByUserAndPetDate(domain.UserID, domain.PetID, time.Time) ([]domain.Report, error)
}

type FindSubscriptionReports struct {
	reportRepo SubscriptionReportRepository
	petRepo    domain.PetRepository
}

func NewFindSubscriptionReports(
	reportRepo SubscriptionReportRepository,
	petRepo domain.PetRepository,
) *FindSubscriptionReports {
	return &FindSubscriptionReports{reportRepo: reportRepo, petRepo: petRepo}
}

func (u *FindSubscriptionReports) Execute(input FindSubscriptionReportsInput) (SubscriptionReportsOutput, error) {
	if !domain.IsValidUserID(input.UserID) || !domain.IsValidPetID(input.PetID) || input.Date.IsZero() {
		return SubscriptionReportsOutput{}, domain.ErrValidation
	}

	pet, err := u.petRepo.FindByID(input.PetID)
	if err != nil {
		return SubscriptionReportsOutput{}, err
	}
	if pet.UserID() != input.UserID {
		return SubscriptionReportsOutput{}, domain.ErrUnauthorized
	}

	reports, err := u.reportRepo.FindByUserAndPetDate(input.UserID, input.PetID, input.Date)
	if err != nil {
		return SubscriptionReportsOutput{}, err
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
