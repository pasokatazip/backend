package persistence

import (
	"database/sql"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type EvolutionRuleRepository struct {
	DB *sql.DB
}

type SatisfiedEvolutionRule struct {
	RuleID        domain.EvolutionRuleID
	ToStageID     domain.EvolutionStageID
	PrimaryStatus string
}

func NewEvolutionRuleRepository(db *sql.DB) *EvolutionRuleRepository {
	return &EvolutionRuleRepository{DB: db}
}

// 迴ｾ蝨ｨ繧ｹ繝・・繧ｸ縺九ｉ菴ｿ縺医ｋ騾ｲ蛹悶Ν繝ｼ繝ｫ荳隕ｧ繧貞叙蠕・
func (r *EvolutionRuleRepository) FindByFromStageID(fromStageID domain.EvolutionStageID) ([]domain.EvolutionRule, error) {
	rows, err := r.DB.Query(
		`SELECT
			id,
			from_stage_id,
			to_stage_id,
			required_experience,
			required_days_since_last_evolution,
			required_feed_count,
			appearance_part,
			created_at,
			updated_at
		FROM evolution_rules
		WHERE from_stage_id = $1
		ORDER BY required_experience, required_feed_count, id`,
		fromStageID,
	)
	if err != nil {
		return nil, mapPersistenceError(err)
	}
	defer rows.Close()

	return scanEvolutionRules(rows)
}

// 謚慕ｨｿ蠕後・邨碁ｨ灘､繝ｻfeed蝗樊焚繝ｻ蜑榊屓騾ｲ蛹匁律繧定ｦ九※縲・ｲ蛹門庄閭ｽ縺ｪ繝ｫ繝ｼ繝ｫ繧・莉ｶ蜿門ｾ・
func (r *EvolutionRuleRepository) FindSatisfiedAfterFeedTx(tx *sql.Tx, petID domain.PetID, checkedAt time.Time) (*SatisfiedEvolutionRule, error) {
	row := tx.QueryRow(
		`WITH pet_snapshot AS (
			SELECT
				p.id,
				p.current_stage_id,
				p.created_at,
				p.curiosity,
				p.sociality,
				p.routine,
				pe.total_experience,
				pe.feed_count,
				CASE
					WHEN ps_status.primary_status = 'curiosity' THEN 'shokushu'
					WHEN ps_status.primary_status = 'sociality' THEN 'yonshoku'
					ELSE 'nishoku'
				END AS branch_key,
				ps_status.primary_status,
				COALESCE(
					(SELECT MAX(evolved_at) FROM pet_evolutions WHERE pet_id = p.id),
					p.created_at
				) AS last_evolved_at
			FROM pets p
			INNER JOIN pet_experiences pe ON pe.pet_id = p.id
			CROSS JOIN LATERAL (
				SELECT CASE
					WHEN p.curiosity >= p.sociality AND p.curiosity >= p.routine THEN 'curiosity'
					WHEN p.sociality >= p.routine THEN 'sociality'
					ELSE 'routine'
				END AS primary_status
			) ps_status
			WHERE p.id = $1
			FOR UPDATE
		)
		SELECT
			er.id,
			er.to_stage_id,
			ps.primary_status
		FROM evolution_rules er
		INNER JOIN pet_snapshot ps ON ps.current_stage_id = er.from_stage_id
		INNER JOIN evolution_stages to_stage ON to_stage.id = er.to_stage_id
		WHERE ps.total_experience >= er.required_experience
		AND ps.feed_count >= er.required_feed_count
		AND DATE_PART('day', $2::timestamptz - ps.last_evolved_at) >= er.required_days_since_last_evolution
		AND (
			to_stage.branch_key IS NULL
			OR to_stage.branch_key = ps.branch_key
			OR er.from_stage_id <> 0
		)
		ORDER BY er.required_experience DESC, er.required_feed_count DESC, er.id
		LIMIT 1`,
		petID,
		checkedAt,
	)

	var (
		ruleID        int
		toStageID     int
		primaryStatus string
	)
	if err := row.Scan(&ruleID, &toStageID, &primaryStatus); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, mapPersistenceError(err)
	}

	return &SatisfiedEvolutionRule{
		RuleID:        domain.EvolutionRuleID(ruleID),
		ToStageID:     domain.EvolutionStageID(toStageID),
		PrimaryStatus: primaryStatus,
	}, nil
}

type evolutionRuleScanner interface {
	Scan(dest ...any) error
}

// SQL縺ｮ蜿門ｾ礼ｵ先棡繧脱volutionRule domain縺ｸ螟画鋤
func scanEvolutionRule(scanner evolutionRuleScanner) (domain.EvolutionRule, error) {
	var (
		id                             int
		fromStageID                    int
		toStageID                      int
		requiredExperience             int64
		requiredDaysSinceLastEvolution int
		requiredFeedCount              int
		appearancePart                 sql.NullString
		createdAt                      time.Time
		updatedAt                      time.Time
	)

	if err := scanner.Scan(
		&id,
		&fromStageID,
		&toStageID,
		&requiredExperience,
		&requiredDaysSinceLastEvolution,
		&requiredFeedCount,
		&appearancePart,
		&createdAt,
		&updatedAt,
	); err != nil {
		return domain.EvolutionRule{}, mapPersistenceError(err)
	}

	var appearancePartValue *string
	if appearancePart.Valid {
		appearancePartValue = &appearancePart.String
	}

	return domain.NewEvolutionRule(
		domain.EvolutionRuleID(id),
		domain.EvolutionStageID(fromStageID),
		domain.EvolutionStageID(toStageID),
		requiredExperience,
		requiredDaysSinceLastEvolution,
		requiredFeedCount,
		appearancePartValue,
		createdAt,
		updatedAt,
	), nil
}

// 隍・焚陦後・SQL邨先棡繧脱volutionRule domain縺ｮ驟榊・縺ｸ螟画鋤
func scanEvolutionRules(rows *sql.Rows) ([]domain.EvolutionRule, error) {
	rules := make([]domain.EvolutionRule, 0)
	for rows.Next() {
		rule, err := scanEvolutionRule(rows)
		if err != nil {
			return nil, mapPersistenceError(err)
		}
		rules = append(rules, rule)
	}
	if err := rows.Err(); err != nil {
		return nil, mapPersistenceError(err)
	}

	return rules, nil
}
