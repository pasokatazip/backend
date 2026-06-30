package persistence

import (
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

// ペットの進化履歴を新規作成
func (r *PetEvolutionRepository) Create(petEvolution domain.PetEvolution) (domain.PetEvolution, error) {
	if err := r.create(r.DB, petEvolution); err != nil {
		return domain.PetEvolution{}, err
	}

	return petEvolution, nil
}

// 指定ペットの進化履歴一覧を新しい順で取得
func (r *PetEvolutionRepository) FindByPetID(petID domain.PetID) ([]domain.PetEvolution, error) {
	rows, err := r.DB.Query(
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
		return nil, err
	}
	defer rows.Close()

	return scanPetEvolutions(rows)
}

// 指定ペットの最新の進化履歴を1件取得
func (r *PetEvolutionRepository) FindLatestByPetID(petID domain.PetID) (domain.PetEvolution, error) {
	row := r.DB.QueryRow(
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

// 進化条件を満たしたルールを適用し、現在ステージ更新 進化履歴作成
func (r *PetEvolutionRepository) ApplySatisfiedRuleTx(tx *sql.Tx, petID domain.PetID, rule SatisfiedEvolutionRule, evolvedAt time.Time) error {
	_, err := tx.Exec(
		`UPDATE pets
		SET
			current_stage_id = $1,
			updated_at = $2
		WHERE id = $3`,
		rule.ToStageID,
		evolvedAt,
		petID,
	)
	if err != nil {
		return err
	}

	primaryStatus := rule.PrimaryStatus
	petEvolution := domain.NewPetEvolution(
		domain.NewPetEvolutionID(),
		petID,
		rule.ToStageID,
		&rule.RuleID,
		&primaryStatus,
		evolvedAt,
		evolvedAt,
	)

	return r.create(tx, petEvolution)
}

type petEvolutionExecer interface {
	Exec(query string, args ...any) (sql.Result, error)
}

// DBまたはトランザクションに対して進化履歴をINSERTする共通処理
func (r *PetEvolutionRepository) create(execer petEvolutionExecer, petEvolution domain.PetEvolution) error {
	_, err := execer.Exec(
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
	return err
}

type petEvolutionScanner interface {
	Scan(dest ...any) error
}

// SQLの取得結果をPetEvolution domainへ変換
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
		return domain.PetEvolution{}, err
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

// 複数行のSQL結果をPetEvolution domainの配列へ変換
func scanPetEvolutions(rows *sql.Rows) ([]domain.PetEvolution, error) {
	evolutions := make([]domain.PetEvolution, 0)
	for rows.Next() {
		evolution, err := scanPetEvolution(rows)
		if err != nil {
			return nil, err
		}
		evolutions = append(evolutions, evolution)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return evolutions, nil
}
