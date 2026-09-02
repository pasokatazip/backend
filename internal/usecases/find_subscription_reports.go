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
	Reports    []ReportOutput
	Pet        SubscriptionReportPetOutput
	HasPraised bool
}

// SubscriptionReportPetOutput contains the presentation data needed to render
// the pet attached to a report. Database IDs for evolution stages are kept out
// of the API contract because the frontend selects animations by stage key.
type SubscriptionReportPetOutput struct {
	ID              string
	Name            string
	Color           string
	CurrentStageKey string
	CurrentStageNo  int
	IsDeleted       bool
	CreatedAt       time.Time
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
	stageRepo  domain.EvolutionStageRepository
	praiseRepo domain.SouvenirPraiseFlagRepository
}

func NewFindSubscriptionReports(
	reportRepo SubscriptionReportRepository,
	petRepo SubscriptionReportPetRepository,
	stageRepo domain.EvolutionStageRepository,
	praiseRepo domain.SouvenirPraiseFlagRepository,
) *FindSubscriptionReports {
	return &FindSubscriptionReports{
		reportRepo: reportRepo,
		petRepo:    petRepo,
		stageRepo:  stageRepo,
		praiseRepo: praiseRepo,
	}
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
	currentStage, err := u.stageRepo.FindByID(domain.EvolutionStageID(pet.CurrentStageID()))
	if err != nil {
		return SubscriptionReportsOutput{}, err
	}

	praiseFlag, err := u.praiseRepo.FindByPetIDAndDate(petID, input.Date)
	if err != nil {
		return SubscriptionReportsOutput{}, err
	}

	reportOutputs := make([]ReportOutput, 0, len(reports))
	for _, report := range reports {
		reportOutputs = append(reportOutputs, reportOutput(report))
	}

	return SubscriptionReportsOutput{
		Reports:    reportOutputs,
		HasPraised: praiseFlag.HasPraised(),
		Pet: SubscriptionReportPetOutput{
			ID:              string(pet.ID()),
			Name:            pet.Name(),
			Color:           pet.Color(),
			CurrentStageKey: currentStage.StageKey(),
			CurrentStageNo:  currentStage.StageNo(),
			IsDeleted:       pet.IsDeleted(),
			CreatedAt:       pet.CreatedAt(),
		},
	}, nil
}
