package persistence

import (
	"database/sql"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/timeutil"
)

type PetExperienceRepository struct {
	DB *sql.DB
}

func NewPetExperienceRepository(db *sql.DB) *PetExperienceRepository {
	return &PetExperienceRepository{DB: db}
}

func (r *PetExperienceRepository) Create(petExperience domain.PetExperience) (domain.PetExperience, error) {
	_, err := r.DB.Exec(
		`INSERT INTO pet_experiences (
			id,
			pet_id,
			total_experience,
			feed_count,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)`,
		petExperience.ID(),
		petExperience.PetID(),
		petExperience.TotalExperience(),
		petExperience.FeedCount(),
		petExperience.CreatedAt(),
		petExperience.UpdatedAt(),
	)
	if err != nil {
		return domain.PetExperience{}, mapPersistenceError(err)
	}

	return petExperience, nil
}

func (r *PetExperienceRepository) FindByPetID(petID domain.PetID) (domain.PetExperience, error) {
	row := r.DB.QueryRow(
		`SELECT
			id,
			pet_id,
			total_experience,
			feed_count,
			created_at,
			updated_at
		FROM pet_experiences
		WHERE pet_id = $1`,
		petID,
	)

	return scanPetExperience(row)
}

func (r *PetExperienceRepository) Update(petExperience domain.PetExperience) (domain.PetExperience, error) {
	_, err := r.DB.Exec(
		`UPDATE pet_experiences
		SET
			total_experience = $1,
			feed_count = $2,
			updated_at = $3
		WHERE id = $4`,
		petExperience.TotalExperience(),
		petExperience.FeedCount(),
		petExperience.UpdatedAt(),
		petExperience.ID(),
	)
	if err != nil {
		return domain.PetExperience{}, mapPersistenceError(err)
	}

	return petExperience, nil
}

type experienceCapUsage struct {
	maxExperience  int
	usedExperience int
}

// calculateAwardedExperience は、有効な取得上限をすべて適用する。
// 日次・週次など複数の上限が重なる場合は、残り取得可能量が最も少ない上限を優先する。
func calculateAwardedExperience(amount int, usages []experienceCapUsage) (awarded int, capped int) {
	if amount <= 0 {
		return 0, 0
	}

	awarded = amount
	for _, usage := range usages {
		remaining := usage.maxExperience - usage.usedExperience
		if remaining < 0 {
			remaining = 0
		}
		if remaining < awarded {
			awarded = remaining
		}
	}

	return awarded, amount - awarded
}

// AddFeedExperienceTx は、有効な取得上限を適用して投稿時の経験値を加算する。
// 経験値が全量制限されても投稿回数は加算し、行ロックによって同時投稿による上限超過を防ぐ。
func (r *PetExperienceRepository) AddFeedExperienceTx(tx *sql.Tx, petID domain.PetID, amount int, occurredAt time.Time) (int, int, error) {
	var lockedPetID string
	if err := tx.QueryRow(
		`SELECT pet_id FROM pet_experiences WHERE pet_id = $1 FOR UPDATE`,
		petID,
	).Scan(&lockedPetID); err != nil {
		return 0, 0, mapPersistenceError(err)
	}

	rows, err := tx.Query(
		`SELECT cap_type, MIN(max_experience)
		FROM experience_caps
		WHERE active = TRUE
		  AND cap_type IN ('daily', 'weekly', 'monthly')
		GROUP BY cap_type`,
	)
	if err != nil {
		return 0, 0, mapPersistenceError(err)
	}

	activeCaps := make(map[domain.ExperienceCapType]int)
	for rows.Next() {
		var capType domain.ExperienceCapType
		var maxExperience int
		if err := rows.Scan(&capType, &maxExperience); err != nil {
			rows.Close()
			return 0, 0, mapPersistenceError(err)
		}
		activeCaps[capType] = maxExperience
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return 0, 0, mapPersistenceError(err)
	}
	rows.Close()

	experienceDate := occurredAt.In(timeutil.LocationJST())
	experienceDate = time.Date(
		experienceDate.Year(), experienceDate.Month(), experienceDate.Day(),
		0, 0, 0, 0, timeutil.LocationJST(),
	)
	usages := make([]experienceCapUsage, 0, len(activeCaps))
	for capType, maxExperience := range activeCaps {
		periodStart := experiencePeriodStart(experienceDate, capType)
		var usedExperience int
		if err := tx.QueryRow(
			`SELECT COALESCE(SUM(GREATEST(amount - capped_amount, 0)), 0)
			FROM pet_experience_events
			WHERE pet_id = $1
			  AND experience_date BETWEEN $2 AND $3`,
			petID,
			periodStart.Format("2006-01-02"),
			experienceDate.Format("2006-01-02"),
		).Scan(&usedExperience); err != nil {
			return 0, 0, mapPersistenceError(err)
		}
		usages = append(usages, experienceCapUsage{
			maxExperience:  maxExperience,
			usedExperience: usedExperience,
		})
	}

	awardedAmount, cappedAmount := calculateAwardedExperience(amount, usages)
	_, err = tx.Exec(
		`UPDATE pet_experiences
		SET total_experience = total_experience + $1,
			feed_count = feed_count + 1,
			updated_at = $2
		WHERE pet_id = $3`,
		awardedAmount,
		occurredAt,
		petID,
	)
	if err != nil {
		return 0, 0, mapPersistenceError(err)
	}

	return awardedAmount, cappedAmount, nil
}

func experiencePeriodStart(experienceDate time.Time, capType domain.ExperienceCapType) time.Time {
	switch capType {
	case domain.ExperienceCapTypeWeekly:
		daysSinceMonday := (int(experienceDate.Weekday()) + 6) % 7
		return experienceDate.AddDate(0, 0, -daysSinceMonday)
	case domain.ExperienceCapTypeMonthly:
		return time.Date(
			experienceDate.Year(), experienceDate.Month(), 1,
			0, 0, 0, 0, experienceDate.Location(),
		)
	default:
		return experienceDate
	}
}

type petExperienceScanner interface {
	Scan(dest ...any) error
}

func scanPetExperience(scanner petExperienceScanner) (domain.PetExperience, error) {
	var (
		id              string
		petID           string
		totalExperience int64
		feedCount       int
		createdAt       time.Time
		updatedAt       time.Time
	)

	if err := scanner.Scan(
		&id,
		&petID,
		&totalExperience,
		&feedCount,
		&createdAt,
		&updatedAt,
	); err != nil {
		return domain.PetExperience{}, mapPersistenceError(err)
	}

	return domain.NewPetExperience(
		domain.PetExperienceID(id),
		domain.PetID(petID),
		totalExperience,
		feedCount,
		createdAt,
		updatedAt,
	), nil
}
