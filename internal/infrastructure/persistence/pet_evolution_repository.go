package persistence

import (
	"context"
	"database/sql"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type PetEvolutionRepository struct {
	DB *sql.DB
}

func NewPetEvolutionRepository(db *sql.DB) *PetEvolutionRepository {
	return &PetEvolutionRepository{DB: db}
}

// Create はペットの進化履歴を新規作成する。
func (r *PetEvolutionRepository) Create(ctx context.Context, petEvolution domain.PetEvolution) (domain.PetEvolution, error) {
	if err := createPetEvolution(ctx, r.DB, petEvolution); err != nil {
		return domain.PetEvolution{}, mapPersistenceError(err)
	}

	return petEvolution, nil
}

// FindByPetID は指定したペットの進化履歴を新しい順に取得する。
func (r *PetEvolutionRepository) FindByPetID(ctx context.Context, petID domain.PetID) ([]domain.PetEvolution, error) {
	rows, err := r.DB.QueryContext(ctx,
		`SELECT
			id,
			pet_id,
			stage_id,
			evolution_rule_id,
			primary_status,
			evolved_at,
			created_at
		FROM pet_evolutions
		WHERE pet_id = $1
		ORDER BY evolved_at DESC`,
		petID,
	)
	if err != nil {
		return nil, mapPersistenceError(err)
	}
	defer rows.Close()

	return scanPetEvolutions(rows)
}

// FindLatestByPetID は指定したペットの最新の進化履歴を取得する。
func (r *PetEvolutionRepository) FindLatestByPetID(ctx context.Context, petID domain.PetID) (domain.PetEvolution, error) {
	row := r.DB.QueryRowContext(ctx,
		`SELECT
			id,
			pet_id,
			stage_id,
			evolution_rule_id,
			primary_status,
			evolved_at,
			created_at
		FROM pet_evolutions
		WHERE pet_id = $1
		ORDER BY evolved_at DESC
		LIMIT 1`,
		petID,
	)

	return scanPetEvolution(row)
}

type petEvolutionExecer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// create はDBまたはトランザクションを使用して進化履歴を登録する。
func createPetEvolution(ctx context.Context, execer petEvolutionExecer, petEvolution domain.PetEvolution) error {
	_, err := execer.ExecContext(ctx,
		`INSERT INTO pet_evolutions (
			id,
			pet_id,
			stage_id,
			evolution_rule_id,
			primary_status,
			evolved_at,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		petEvolution.ID(),
		petEvolution.PetID(),
		petEvolution.StageID(),
		petEvolution.EvolutionRuleID(),
		petEvolution.PrimaryStatus(),
		petEvolution.EvolvedAt(),
		petEvolution.CreatedAt(),
	)
	return mapPersistenceError(err)
}

type petEvolutionScanner interface {
	Scan(dest ...any) error
}

// scanPetEvolution はSQLの取得結果をPetEvolutionドメインへ変換する。
func scanPetEvolution(scanner petEvolutionScanner) (domain.PetEvolution, error) {
	var (
		id              string
		petID           string
		stageID         int
		evolutionRuleID sql.NullInt64
		primaryStatus   sql.NullString
		evolvedAt       time.Time
		createdAt       time.Time
	)

	if err := scanner.Scan(
		&id,
		&petID,
		&stageID,
		&evolutionRuleID,
		&primaryStatus,
		&evolvedAt,
		&createdAt,
	); err != nil {
		return domain.PetEvolution{}, mapPersistenceError(err)
	}

	var evolutionRuleIDValue *domain.EvolutionRuleID
	if evolutionRuleID.Valid {
		value := domain.EvolutionRuleID(evolutionRuleID.Int64)
		evolutionRuleIDValue = &value
	}

	var primaryStatusValue *string
	if primaryStatus.Valid {
		primaryStatusValue = &primaryStatus.String
	}

	return domain.NewPetEvolution(
		domain.PetEvolutionID(id),
		domain.PetID(petID),
		domain.EvolutionStageID(stageID),
		evolutionRuleIDValue,
		primaryStatusValue,
		evolvedAt,
		createdAt,
	), nil
}

// scanPetEvolutions は複数行のSQL結果をPetEvolutionドメインの配列へ変換する。
func scanPetEvolutions(rows *sql.Rows) ([]domain.PetEvolution, error) {
	evolutions := make([]domain.PetEvolution, 0)
	for rows.Next() {
		evolution, err := scanPetEvolution(rows)
		if err != nil {
			return nil, mapPersistenceError(err)
		}
		evolutions = append(evolutions, evolution)
	}
	if err := rows.Err(); err != nil {
		return nil, mapPersistenceError(err)
	}

	return evolutions, nil
}

// CreateInTransaction は呼び出し元のトランザクション内で進化履歴を保存する。
func (r *PetEvolutionRepository) CreateInTransaction(ctx context.Context, evolution domain.PetEvolution) error {
	tx, err := requireTransaction(ctx, r.DB)
	if err != nil {
		return err
	}
	return createPetEvolution(ctx, tx, evolution)
}
