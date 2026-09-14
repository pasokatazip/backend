package usecases

import (
	"context"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/timeutil"
)

type SouvenirPraiseFlagOutput struct {
	HasPraised bool
	ReportDate time.Time
	PraisedAt  *time.Time
}

type MarkSouvenirPraised struct {
	praiseRepo domain.SouvenirPraiseFlagRepository
	reportRepo SubscriptionReportRepository
}

func NewMarkSouvenirPraised(
	praiseRepo domain.SouvenirPraiseFlagRepository,
	reportRepo SubscriptionReportRepository,
) *MarkSouvenirPraised {
	return &MarkSouvenirPraised{praiseRepo: praiseRepo, reportRepo: reportRepo}
}

type MarkSouvenirPraisedInput struct {
	UserID     domain.UserID
	ReportDate time.Time
}

func (u *MarkSouvenirPraised) Execute(ctx context.Context,
	input MarkSouvenirPraisedInput,
) (SouvenirPraiseFlagOutput, error) {
	if !domain.IsValidUserID(input.UserID) || input.ReportDate.IsZero() {
		return SouvenirPraiseFlagOutput{}, domain.ErrValidation
	}

	reportDate := normalizeJSTDate(input.ReportDate)
	if reportDate.After(normalizeJSTDate(timeutil.NowJST())) {
		return SouvenirPraiseFlagOutput{}, domain.ErrValidation
	}

	reports, err := u.reportRepo.FindByUserAndDate(ctx, input.UserID, reportDate)
	if err != nil {
		return SouvenirPraiseFlagOutput{}, err
	}
	if len(reports) == 0 {
		return SouvenirPraiseFlagOutput{}, domain.ErrNotFound
	}

	hasSouvenir := false
	for _, report := range reports {
		if len(report.Souvenirs()) > 0 {
			hasSouvenir = true
			break
		}
	}
	if !hasSouvenir {
		return SouvenirPraiseFlagOutput{}, domain.ErrNotFound
	}

	flag, err := u.praiseRepo.MarkPraised(ctx, input.UserID, reportDate)
	if err != nil {
		return SouvenirPraiseFlagOutput{}, err
	}
	return souvenirPraiseFlagOutput(flag), nil
}

func normalizeJSTDate(value time.Time) time.Time {
	dateInJST := value.In(timeutil.LocationJST())
	return time.Date(
		dateInJST.Year(), dateInJST.Month(), dateInJST.Day(),
		0, 0, 0, 0, timeutil.LocationJST(),
	)
}

func souvenirPraiseFlagOutput(flag domain.SouvenirPraiseFlag) SouvenirPraiseFlagOutput {
	return SouvenirPraiseFlagOutput{
		HasPraised: flag.HasPraised(),
		ReportDate: flag.ReportDate(),
		PraisedAt:  flag.PraisedAt(),
	}
}
