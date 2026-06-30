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

// 経験値取得イベントを新規作成
func (r *PetExperienceEventRepository) Create(petExperienceEvent domain.PetExperienceEvent) (domain.PetExperienceEvent, error) {
	if err := r.create(r.DB, petExperienceEvent); err != nil {
		return domain.PetExperienceEvent{}, err
	}

	return petExperienceEvent, nil
}

// 経験値取得イベントを同一トランザクション内で作成
func (r *PetExperienceEventRepository) CreateTx(tx *sql.Tx, petExperienceEvent domain.PetExperienceEvent) error {
	return r.create(tx, petExperienceEvent)
}

// 指定ペットの経験値取得イベント一覧を新しい順で取得
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
		return nil, err
	}
	defer rows.Close()

	return scanPetExperienceEvents(rows)
}

// 指定ペットの指定日の経験値取得イベント一覧を取得
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
		return nil, err
	}
	defer rows.Close()

	return scanPetExperienceEvents(rows)
}

type petExperienceEventExecer interface {
	Exec(query string, args ...any) (sql.Result, error)
}

// DBまたはトランザクションに対して経験値取得イベントをINSERTする共通処理
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
	return err
}

type petExperienceEventScanner interface {
	Scan(dest ...any) error
}

// SQLの取得結果をPetExperienceEvent domainへ変換
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
		return domain.PetExperienceEvent{}, err
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

// 複数行のSQL結果をPetExperienceEvent domainの配列へ変換
func scanPetExperienceEvents(rows *sql.Rows) ([]domain.PetExperienceEvent, error) {
	events := make([]domain.PetExperienceEvent, 0)
	for rows.Next() {
		event, err := scanPetExperienceEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}
