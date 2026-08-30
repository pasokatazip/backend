package domain

import "time"

// SouvenirPraiseFlag represents whether a user has completed the praise
// interaction for one report date.
type SouvenirPraiseFlag struct {
	userID     UserID
	reportDate time.Time
	hasPraised bool
	praisedAt  *time.Time
}

func NewSouvenirPraiseFlag(
	userID UserID,
	reportDate time.Time,
	hasPraised bool,
	praisedAt *time.Time,
) SouvenirPraiseFlag {
	return SouvenirPraiseFlag{
		userID:     userID,
		reportDate: reportDate,
		hasPraised: hasPraised,
		praisedAt:  praisedAt,
	}
}

func (f SouvenirPraiseFlag) UserID() UserID        { return f.userID }
func (f SouvenirPraiseFlag) ReportDate() time.Time { return f.reportDate }
func (f SouvenirPraiseFlag) HasPraised() bool      { return f.hasPraised }
func (f SouvenirPraiseFlag) PraisedAt() *time.Time { return f.praisedAt }

type SouvenirPraiseFlagRepository interface {
	// FindByPetIDAndDate resolves the pet owner and returns an unpraised flag
	// when no row exists for that report date.
	FindByPetIDAndDate(petID PetID, reportDate time.Time) (SouvenirPraiseFlag, error)
	// MarkPraised records the first praise time for a report date. Repeated
	// calls for the same user and date are idempotent.
	MarkPraised(userID UserID, reportDate time.Time) (SouvenirPraiseFlag, error)
}
