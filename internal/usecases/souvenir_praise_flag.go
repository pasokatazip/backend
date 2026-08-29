package usecases

import (
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type SouvenirPraiseFlagOutput struct {
	HasPraised bool
	ReportDate time.Time
	PraisedAt  *time.Time
}

type MarkSouvenirPraised struct {
	repo domain.SouvenirPraiseFlagRepository
}

func NewMarkSouvenirPraised(
	repo domain.SouvenirPraiseFlagRepository,
) *MarkSouvenirPraised {
	return &MarkSouvenirPraised{repo: repo}
}

type MarkSouvenirPraisedInput struct {
	UserID     domain.UserID
	ReportDate time.Time
}

func (u *MarkSouvenirPraised) Execute(
	input MarkSouvenirPraisedInput,
) (SouvenirPraiseFlagOutput, error) {
	if !domain.IsValidUserID(input.UserID) || input.ReportDate.IsZero() {
		return SouvenirPraiseFlagOutput{}, domain.ErrValidation
	}

	flag, err := u.repo.MarkPraised(input.UserID, input.ReportDate)
	if err != nil {
		return SouvenirPraiseFlagOutput{}, err
	}
	return souvenirPraiseFlagOutput(flag), nil
}

func souvenirPraiseFlagOutput(flag domain.SouvenirPraiseFlag) SouvenirPraiseFlagOutput {
	return SouvenirPraiseFlagOutput{
		HasPraised: flag.HasPraised(),
		ReportDate: flag.ReportDate(),
		PraisedAt:  flag.PraisedAt(),
	}
}
