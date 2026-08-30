package persistence

import (
	"database/sql"
	"errors"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/timeutil"
)

type SouvenirPraiseFlagRepository struct {
	DB *sql.DB
}

func NewSouvenirPraiseFlagRepository(db *sql.DB) *SouvenirPraiseFlagRepository {
	return &SouvenirPraiseFlagRepository{DB: db}
}

func (r *SouvenirPraiseFlagRepository) FindByPetIDAndDate(
	petID domain.PetID,
	reportDate time.Time,
) (domain.SouvenirPraiseFlag, error) {
	row := r.DB.QueryRow(
		`SELECT
			p.user_id,
			$2::date,
			COALESCE(flag.has_praised, FALSE),
			flag.praised_at
		 FROM pets p
		 LEFT JOIN user_souvenir_praise_flags flag
			ON flag.user_id = p.user_id
			AND flag.report_date = $2::date
		 WHERE p.id = $1`,
		petID,
		praiseReportDateValue(reportDate),
	)

	flag, err := scanSouvenirPraiseFlag(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.SouvenirPraiseFlag{}, domain.ErrNotFound
		}
		return domain.SouvenirPraiseFlag{}, mapPersistenceError(err)
	}
	return flag, nil
}

func (r *SouvenirPraiseFlagRepository) MarkPraised(
	userID domain.UserID,
	reportDate time.Time,
) (domain.SouvenirPraiseFlag, error) {
	row := r.DB.QueryRow(
		`INSERT INTO user_souvenir_praise_flags (
			user_id, report_date, has_praised, praised_at
		 ) VALUES ($1, $2::date, TRUE, CURRENT_TIMESTAMP)
		 ON CONFLICT (user_id, report_date) DO UPDATE
		 SET
			has_praised = TRUE,
			praised_at = COALESCE(
				user_souvenir_praise_flags.praised_at,
				EXCLUDED.praised_at
			),
			updated_at = CURRENT_TIMESTAMP
		 RETURNING user_id, report_date, has_praised, praised_at`,
		userID,
		praiseReportDateValue(reportDate),
	)

	flag, err := scanSouvenirPraiseFlag(row)
	if err != nil {
		return domain.SouvenirPraiseFlag{}, mapPersistenceError(err)
	}
	return flag, nil
}

func praiseReportDateValue(reportDate time.Time) string {
	return reportDate.In(timeutil.LocationJST()).Format("2006-01-02")
}

type souvenirPraiseFlagScanner interface {
	Scan(dest ...any) error
}

func scanSouvenirPraiseFlag(scanner souvenirPraiseFlagScanner) (domain.SouvenirPraiseFlag, error) {
	var (
		userID     string
		reportDate time.Time
		hasPraised bool
		praisedAt  sql.NullTime
	)
	if err := scanner.Scan(&userID, &reportDate, &hasPraised, &praisedAt); err != nil {
		return domain.SouvenirPraiseFlag{}, err
	}

	var praisedAtValue *time.Time
	if praisedAt.Valid {
		praisedAtValue = &praisedAt.Time
	}
	return domain.NewSouvenirPraiseFlag(
		domain.UserID(userID),
		reportDate,
		hasPraised,
		praisedAtValue,
	), nil
}
