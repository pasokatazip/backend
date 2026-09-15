package persistence

import (
	"context"
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

// Create は経験値取得イベントを新規作成する。
func (r *PetExperienceEventRepository) Create(ctx context.Context, petExperienceEvent domain.PetExperienceEvent) (domain.PetExperienceEvent, error) {
	if err := createPetExperienceEvent(ctx, r.DB, petExperienceEvent); err != nil {
		return domain.PetExperienceEvent{}, mapPersistenceError(err)
	}

	return petExperienceEvent, nil
}

// FindByPetID は指定したペットの経験値取得イベントを新しい順に取得する。
func (r *PetExperienceEventRepository) FindByPetID(ctx context.Context, petID domain.PetID) ([]domain.PetExperienceEvent, error) {
	rows, err := r.DB.QueryContext(ctx,
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

// FindByPetIDAndDate は指定したペットと日付の経験値取得イベントを取得する。
func (r *PetExperienceEventRepository) FindByPetIDAndDate(ctx context.Context, petID domain.PetID, experienceDate time.Time) ([]domain.PetExperienceEvent, error) {
	rows, err := r.DB.QueryContext(ctx,
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
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// create はDBまたはトランザクションを使用して経験値取得イベントを登録する。
func createPetExperienceEvent(ctx context.Context, execer petExperienceEventExecer, petExperienceEvent domain.PetExperienceEvent) error {
	_, err := execer.ExecContext(ctx,
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

// scanPetExperienceEvent はSQLの取得結果をPetExperienceEventドメインへ変換する。
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

// scanPetExperienceEvents は複数行のSQL結果をPetExperienceEventドメインの配列へ変換する。
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

// CreateInTransaction は呼び出し元のトランザクション内で経験値取得イベントを保存する。
func (r *PetExperienceEventRepository) CreateInTransaction(ctx context.Context, event domain.PetExperienceEvent) error {
	tx, err := requireTransaction(ctx, r.DB)
	if err != nil {
		return err
	}
	return createPetExperienceEvent(ctx, tx, event)
}
