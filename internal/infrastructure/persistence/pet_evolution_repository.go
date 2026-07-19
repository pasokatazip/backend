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

// 繝壹ャ繝医・騾ｲ蛹門ｱ･豁ｴ繧呈眠隕丈ｽ懈・
func (r *PetEvolutionRepository) Create(petEvolution domain.PetEvolution) (domain.PetEvolution, error) {
	if err := r.create(r.DB, petEvolution); err != nil {
		return domain.PetEvolution{}, mapPersistenceError(err)
	}

	return petEvolution, nil
}

// 謖・ｮ壹・繝・ヨ縺ｮ騾ｲ蛹門ｱ･豁ｴ荳隕ｧ繧呈眠縺励＞鬆・〒蜿門ｾ・
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
		return nil, mapPersistenceError(err)
	}
	defer rows.Close()

	return scanPetEvolutions(rows)
}

// 謖・ｮ壹・繝・ヨ縺ｮ譛譁ｰ縺ｮ騾ｲ蛹門ｱ･豁ｴ繧・莉ｶ蜿門ｾ・
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

// 騾ｲ蛹匁擅莉ｶ繧呈ｺ縺溘＠縺溘Ν繝ｼ繝ｫ繧帝←逕ｨ縺励∫樟蝨ｨ繧ｹ繝・・繧ｸ譖ｴ譁ｰ 騾ｲ蛹門ｱ･豁ｴ菴懈・
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
		return mapPersistenceError(err)
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

// DB縺ｾ縺溘・繝医Λ繝ｳ繧ｶ繧ｯ繧ｷ繝ｧ繝ｳ縺ｫ蟇ｾ縺励※騾ｲ蛹門ｱ･豁ｴ繧棚NSERT縺吶ｋ蜈ｱ騾壼・逅・
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
	return mapPersistenceError(err)
}

type petEvolutionScanner interface {
	Scan(dest ...any) error
}

// SQL縺ｮ蜿門ｾ礼ｵ先棡繧単etEvolution domain縺ｸ螟画鋤
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

// 隍・焚陦後・SQL邨先棡繧単etEvolution domain縺ｮ驟榊・縺ｸ螟画鋤
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
