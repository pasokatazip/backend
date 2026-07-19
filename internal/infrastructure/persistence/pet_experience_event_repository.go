package persistence

import (
	"database/sql"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type PetExperienceEventRepository struct {
	DB *sql.DB
}

func NewPetExperienceEventRepository(db *sql.DB) *PetExperienceEventRepository {
	return &PetExperienceEventRepository{DB: db}
}

// 邨碁ｨ灘､蜿門ｾ励う繝吶Φ繝医ｒ譁ｰ隕丈ｽ懈・
func (r *PetExperienceEventRepository) Create(petExperienceEvent domain.PetExperienceEvent) (domain.PetExperienceEvent, error) {
	if err := r.create(r.DB, petExperienceEvent); err != nil {
		return domain.PetExperienceEvent{}, mapPersistenceError(err)
	}

	return petExperienceEvent, nil
}

// 邨碁ｨ灘､蜿門ｾ励う繝吶Φ繝医ｒ蜷御ｸ繝医Λ繝ｳ繧ｶ繧ｯ繧ｷ繝ｧ繝ｳ蜀・〒菴懈・
func (r *PetExperienceEventRepository) CreateTx(tx *sql.Tx, petExperienceEvent domain.PetExperienceEvent) error {
	return r.create(tx, petExperienceEvent)
}

// 謖・ｮ壹・繝・ヨ縺ｮ邨碁ｨ灘､蜿門ｾ励う繝吶Φ繝井ｸ隕ｧ繧呈眠縺励＞鬆・〒蜿門ｾ・
func (r *PetExperienceEventRepository) FindByPetID(petID domain.PetID) ([]domain.PetExperienceEvent, error) {
	rows, err := r.DB.Query(
		`SELECT
			id,
			pet_id,
			source_type,
			source_id,
			amount,
			capped_amount,
			experience_date,
			created_at
		FROM pet_experience_events
		WHERE pet_id = $1
		ORDER BY created_at DESC`,
		petID,
	)
	if err != nil {
		return nil, mapPersistenceError(err)
	}
	defer rows.Close()

	return scanPetExperienceEvents(rows)
}

// 謖・ｮ壹・繝・ヨ縺ｮ謖・ｮ壽律縺ｮ邨碁ｨ灘､蜿門ｾ励う繝吶Φ繝井ｸ隕ｧ繧貞叙蠕・
func (r *PetExperienceEventRepository) FindByPetIDAndDate(petID domain.PetID, experienceDate time.Time) ([]domain.PetExperienceEvent, error) {
	rows, err := r.DB.Query(
		`SELECT
			id,
			pet_id,
			source_type,
			source_id,
			amount,
			capped_amount,
			experience_date,
			created_at
		FROM pet_experience_events
		WHERE pet_id = $1
		AND experience_date = $2
		ORDER BY created_at DESC`,
		petID,
		experienceDate.Format("2006-01-02"),
	)
	if err != nil {
		return nil, mapPersistenceError(err)
	}
	defer rows.Close()

	return scanPetExperienceEvents(rows)
}

type petExperienceEventExecer interface {
	Exec(query string, args ...any) (sql.Result, error)
}

// DB縺ｾ縺溘・繝医Λ繝ｳ繧ｶ繧ｯ繧ｷ繝ｧ繝ｳ縺ｫ蟇ｾ縺励※邨碁ｨ灘､蜿門ｾ励う繝吶Φ繝医ｒINSERT縺吶ｋ蜈ｱ騾壼・逅・
func (r *PetExperienceEventRepository) create(execer petExperienceEventExecer, petExperienceEvent domain.PetExperienceEvent) error {
	_, err := execer.Exec(
		`INSERT INTO pet_experience_events (
			id,
			pet_id,
			source_type,
			source_id,
			amount,
			capped_amount,
			experience_date,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		petExperienceEvent.ID(),
		petExperienceEvent.PetID(),
		petExperienceEvent.SourceType(),
		petExperienceEvent.SourceID(),
		petExperienceEvent.Amount(),
		petExperienceEvent.CappedAmount(),
		petExperienceEvent.ExperienceDate().Format("2006-01-02"),
		petExperienceEvent.CreatedAt(),
	)
	return mapPersistenceError(err)
}

type petExperienceEventScanner interface {
	Scan(dest ...any) error
}

// SQL縺ｮ蜿門ｾ礼ｵ先棡繧単etExperienceEvent domain縺ｸ螟画鋤
func scanPetExperienceEvent(scanner petExperienceEventScanner) (domain.PetExperienceEvent, error) {
	var (
		id             string
		petID          string
		sourceType     string
		sourceID       sql.NullString
		amount         int
		cappedAmount   int
		experienceDate time.Time
		createdAt      time.Time
	)

	if err := scanner.Scan(
		&id,
		&petID,
		&sourceType,
		&sourceID,
		&amount,
		&cappedAmount,
		&experienceDate,
		&createdAt,
	); err != nil {
		return domain.PetExperienceEvent{}, mapPersistenceError(err)
	}

	var sourceIDValue *string
	if sourceID.Valid {
		sourceIDValue = &sourceID.String
	}

	return domain.NewPetExperienceEvent(
		domain.PetExperienceEventID(id),
		domain.PetID(petID),
		domain.ExperienceSourceType(sourceType),
		sourceIDValue,
		amount,
		cappedAmount,
		experienceDate,
		createdAt,
	), nil
}

// 隍・焚陦後・SQL邨先棡繧単etExperienceEvent domain縺ｮ驟榊・縺ｸ螟画鋤
func scanPetExperienceEvents(rows *sql.Rows) ([]domain.PetExperienceEvent, error) {
	events := make([]domain.PetExperienceEvent, 0)
	for rows.Next() {
		event, err := scanPetExperienceEvent(rows)
		if err != nil {
			return nil, mapPersistenceError(err)
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, mapPersistenceError(err)
	}

	return events, nil
}
