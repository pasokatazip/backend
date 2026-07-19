package persistence

import (
	"database/sql"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type EvolutionStageRepository struct {
	DB *sql.DB
}

func NewEvolutionStageRepository(db *sql.DB) *EvolutionStageRepository {
	return &EvolutionStageRepository{DB: db}
}

func (r *EvolutionStageRepository) FindByID(id domain.EvolutionStageID) (domain.EvolutionStage, error) {
	return scanEvolutionStage(r.DB.QueryRow(
		`SELECT
			id,
			stage_key,
			stage_no,
			name,
			branch_key,
			image_url,
			created_at,
			updated_at
		FROM evolution_stages
		WHERE id = $1`,
		id,
	))
}

func (r *EvolutionStageRepository) FindByStageNo(stageNo int) (domain.EvolutionStage, error) {
	return scanEvolutionStage(r.DB.QueryRow(
		`SELECT
			id,
			stage_key,
			stage_no,
			name,
			branch_key,
			image_url,
			created_at,
			updated_at
		FROM evolution_stages
		WHERE stage_no = $1
		ORDER BY id
		LIMIT 1`,
		stageNo,
	))
}

func (r *EvolutionStageRepository) FindAll() ([]domain.EvolutionStage, error) {
	rows, err := r.DB.Query(
		`SELECT
			id,
			stage_key,
			stage_no,
			name,
			branch_key,
			image_url,
			created_at,
			updated_at
		FROM evolution_stages
		ORDER BY stage_no, id`,
	)
	if err != nil {
		return nil, mapPersistenceError(err)
	}
	defer rows.Close()

	stages := make([]domain.EvolutionStage, 0)
	for rows.Next() {
		stage, err := scanEvolutionStage(rows)
		if err != nil {
			return nil, mapPersistenceError(err)
		}
		stages = append(stages, stage)
	}
	if err := rows.Err(); err != nil {
		return nil, mapPersistenceError(err)
	}

	return stages, nil
}

type evolutionStageScanner interface {
	Scan(dest ...any) error
}

func scanEvolutionStage(scanner evolutionStageScanner) (domain.EvolutionStage, error) {
	var (
		id        int
		stageKey  string
		stageNo   int
		name      string
		branchKey sql.NullString
		imageURL  sql.NullString
		createdAt time.Time
		updatedAt time.Time
	)

	if err := scanner.Scan(
		&id,
		&stageKey,
		&stageNo,
		&name,
		&branchKey,
		&imageURL,
		&createdAt,
		&updatedAt,
	); err != nil {
		return domain.EvolutionStage{}, mapPersistenceError(err)
	}

	var imageURLValue *string
	if imageURL.Valid {
		imageURLValue = &imageURL.String
	}

	var branchKeyValue *string
	if branchKey.Valid {
		branchKeyValue = &branchKey.String
	}

	return domain.NewEvolutionStage(
		domain.EvolutionStageID(id),
		stageKey,
		stageNo,
		name,
		branchKeyValue,
		imageURLValue,
		createdAt,
		updatedAt,
	), nil
}
