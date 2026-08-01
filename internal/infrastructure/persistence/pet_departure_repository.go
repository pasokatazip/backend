package persistence

import (
	"database/sql"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type PetDepartureRepository struct {
	DB *sql.DB
}

func NewPetDepartureRepository(db *sql.DB) *PetDepartureRepository {
	return &PetDepartureRepository{DB: db}
}

func (r *PetDepartureRepository) FindActiveRule() (domain.PetDepartureRule, error) {
	row := r.DB.QueryRow(
		`SELECT
			pdr.id,
			pdr.rule_key,
			pdr.min_age_days,
			pdr.required_stage_id,
			es.stage_no,
			pdr.grace_days_min,
			pdr.grace_days_max
		FROM pet_departure_rules pdr
		INNER JOIN evolution_stages es ON es.id = pdr.required_stage_id
		WHERE pdr.active = TRUE
		ORDER BY pdr.id
		LIMIT 1`,
	)

	var rule domain.PetDepartureRule
	var id int
	if err := row.Scan(
		&id,
		&rule.RuleKey,
		&rule.MinAgeDays,
		&rule.RequiredStageID,
		&rule.RequiredStageNo,
		&rule.GraceDaysMin,
		&rule.GraceDaysMax,
	); err != nil {
		return domain.PetDepartureRule{}, mapPersistenceError(err)
	}
	rule.ID = domain.PetDepartureRuleID(id)

	return rule, nil
}

func (r *PetDepartureRepository) FindActivePetsByUserID(rule domain.PetDepartureRule, userID domain.UserID) ([]domain.PetDepartureCandidate, error) {
	rows, err := r.DB.Query(
		`SELECT
			p.id,
			p.user_id,
			p.created_at,
			p.current_stage_id,
			current_stage.stage_no,
			stage_reached.stage_reached_at,
			pd.id,
			pd.status,
			pd.eligible_at,
			pd.scheduled_departure_at
		FROM pets p
		INNER JOIN user_active_pets uap ON uap.pet_id = p.id
		INNER JOIN evolution_stages current_stage ON current_stage.id = p.current_stage_id
		LEFT JOIN (
			SELECT
				pe.pet_id,
				MIN(pe.evolved_at) AS stage_reached_at
			FROM pet_evolutions pe
			INNER JOIN evolution_stages es ON es.id = pe.stage_id
			WHERE es.stage_no >= $1
			GROUP BY pe.pet_id
		) stage_reached ON stage_reached.pet_id = p.id
		LEFT JOIN pet_departures pd ON pd.pet_id = p.id
		WHERE p.is_deleted = FALSE
			AND p.status = 'active'
			AND uap.user_id = $2
		ORDER BY p.created_at`,
		rule.RequiredStageNo,
		userID,
	)
	if err != nil {
		return nil, mapPersistenceError(err)
	}
	defer rows.Close()

	pets := make([]domain.PetDepartureCandidate, 0)
	for rows.Next() {
		pet, err := scanPetDepartureCandidate(rows)
		if err != nil {
			return nil, mapPersistenceError(err)
		}
		pets = append(pets, pet)
	}
	if err := rows.Err(); err != nil {
		return nil, mapPersistenceError(err)
	}

	return pets, nil
}

func (r *PetDepartureRepository) FindByPetID(petID domain.PetID) (domain.PetDeparture, error) {
	row := r.DB.QueryRow(
		`SELECT status, eligible_at, scheduled_departure_at
		FROM pet_departures
		WHERE pet_id = $1`,
		petID,
	)

	var (
		departure   domain.PetDeparture
		eligibleAt  sql.NullTime
		scheduledAt sql.NullTime
	)
	if err := row.Scan(&departure.Status, &eligibleAt, &scheduledAt); err != nil {
		return domain.PetDeparture{}, mapPersistenceError(err)
	}
	departure.EligibleAt = nullableDepartureTime(eligibleAt)
	departure.ScheduledDepartureAt = nullableDepartureTime(scheduledAt)
	return departure, nil
}

func (r *PetDepartureRepository) Upsert(input domain.PetDepartureUpsertInput) error {
	_, err := r.DB.Exec(
		`INSERT INTO pet_departures (
			id,
			pet_id,
			user_id,
			pet_departure_rule_id,
			eligible_at,
			scheduled_departure_at,
			status,
			blocked_reason,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
		ON CONFLICT (pet_id) DO UPDATE
		SET
			user_id = EXCLUDED.user_id,
			pet_departure_rule_id = EXCLUDED.pet_departure_rule_id,
			eligible_at = EXCLUDED.eligible_at,
			scheduled_departure_at = EXCLUDED.scheduled_departure_at,
			status = EXCLUDED.status,
			blocked_reason = EXCLUDED.blocked_reason,
			updated_at = EXCLUDED.updated_at`,
		domain.NewUUIDString(),
		input.PetID,
		input.UserID,
		input.RuleID,
		input.EligibleAt,
		input.ScheduledDepartureAt,
		input.Status,
		input.BlockedReason,
		input.CheckedAt,
	)
	return mapPersistenceError(err)
}

func (r *PetDepartureRepository) Depart(input domain.PetDepartureDepartInput) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return mapPersistenceError(err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`INSERT INTO pet_departures (
			id,
			pet_id,
			user_id,
			pet_departure_rule_id,
			eligible_at,
			scheduled_departure_at,
			departed_at,
			status,
			blocked_reason,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, 'departed', NULL, $7, $7)
		ON CONFLICT (pet_id) DO UPDATE
		SET
			user_id = EXCLUDED.user_id,
			pet_departure_rule_id = EXCLUDED.pet_departure_rule_id,
			eligible_at = EXCLUDED.eligible_at,
			scheduled_departure_at = EXCLUDED.scheduled_departure_at,
			departed_at = EXCLUDED.departed_at,
			status = EXCLUDED.status,
			blocked_reason = NULL,
			updated_at = EXCLUDED.updated_at`,
		domain.NewUUIDString(),
		input.PetID,
		input.UserID,
		input.RuleID,
		input.EligibleAt,
		input.ScheduledAt,
		input.DepartedAt,
	)
	if err != nil {
		return mapPersistenceError(err)
	}

	_, err = tx.Exec(
		`UPDATE pets
		SET
			status = 'departed',
			updated_at = $1
		WHERE id = $2
			AND status = 'active'`,
		input.DepartedAt,
		input.PetID,
	)
	if err != nil {
		return mapPersistenceError(err)
	}

	_, err = tx.Exec(
		`DELETE FROM user_active_pets
		WHERE user_id = $1
			AND pet_id = $2`,
		input.UserID,
		input.PetID,
	)
	if err != nil {
		return mapPersistenceError(err)
	}

	_, err = tx.Exec(
		`UPDATE pet_group_joins
		SET
			left_at = $1,
			updated_at = $1
		WHERE pet_id = $2
			AND left_at IS NULL`,
		input.DepartedAt,
		input.PetID,
	)
	if err != nil {
		return mapPersistenceError(err)
	}

	return tx.Commit()
}

type petDepartureCandidateScanner interface {
	Scan(dest ...any) error
}

func scanPetDepartureCandidate(scanner petDepartureCandidateScanner) (domain.PetDepartureCandidate, error) {
	var (
		petID                string
		userID               string
		createdAt            time.Time
		currentStageID       int
		currentStageNo       int
		stageReachedAt       sql.NullTime
		departureID          sql.NullString
		departureStatus      sql.NullString
		eligibleAt           sql.NullTime
		scheduledDepartureAt sql.NullTime
	)

	if err := scanner.Scan(
		&petID,
		&userID,
		&createdAt,
		&currentStageID,
		&currentStageNo,
		&stageReachedAt,
		&departureID,
		&departureStatus,
		&eligibleAt,
		&scheduledDepartureAt,
	); err != nil {
		return domain.PetDepartureCandidate{}, mapPersistenceError(err)
	}

	return domain.PetDepartureCandidate{
		PetID:                domain.PetID(petID),
		UserID:               domain.UserID(userID),
		CreatedAt:            createdAt,
		CurrentStageID:       currentStageID,
		CurrentStageNo:       currentStageNo,
		StageReachedAt:       nullableDepartureTime(stageReachedAt),
		DepartureID:          nullableDepartureString(departureID),
		DepartureStatus:      nullableDepartureString(departureStatus),
		EligibleAt:           nullableDepartureTime(eligibleAt),
		ScheduledDepartureAt: nullableDepartureTime(scheduledDepartureAt),
	}, nil
}

func nullableDepartureString(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func nullableDepartureTime(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	return &value.Time
}
