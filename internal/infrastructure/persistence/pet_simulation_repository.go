package persistence

import (
	"database/sql"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type PetSimulationRepository struct {
	DB *sql.DB
}

func NewPetSimulationRepository(db *sql.DB) *PetSimulationRepository {
	return &PetSimulationRepository{DB: db}
}

func (r *PetSimulationRepository) FindActivePetsForSimulation() ([]domain.SimulationPet, error) {
	query := `
		SELECT
			p.id,
			p.name,
			p.color,
			p.is_deleted,
			p.user_id,
			p.energy,
			p.curiosity,
			p.sociality,
			p.routine,
			p.current_group_master_id,
			p.current_stage_id,
			p.created_at,
			p.updated_at,
			pgj.id,
			pgj.joined_at
		FROM pets p
		INNER JOIN user_active_pets uap ON uap.pet_id = p.id
		LEFT JOIN pet_group_joins pgj
			ON pgj.pet_id = p.id
			AND pgj.left_at IS NULL
		WHERE p.is_deleted = FALSE
		ORDER BY p.created_at
	`

	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pets := make([]domain.SimulationPet, 0)
	for rows.Next() {
		pet, currentJoinID, joinedAt, err := scanSimulationPet(rows)
		if err != nil {
			return nil, err
		}
		pets = append(pets, domain.SimulationPet{
			Pet:           pet,
			CurrentJoinID: currentJoinID,
			JoinedAt:      joinedAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return pets, nil
}

func (r *PetSimulationRepository) FindActiveGroupsForSimulation() ([]domain.GroupMaster, error) {
	return NewGroupMasterRepository(r.DB).FindActive()
}

func (r *PetSimulationRepository) SaveHourlySimulation(input domain.PetSimulationSaveInput) (bool, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var exists bool
	// 同じ時刻・ペットのログがあるか確認
	if err := tx.QueryRow(
		`SELECT EXISTS (SELECT 1 FROM pet_hourly_logs WHERE pet_id = $1 AND simulated_at = $2)`,
		input.PetID,
		input.SimulatedAt,
	).Scan(&exists); err != nil {
		return false, err
	}
	if exists {
		return false, nil
	}

	var newJoinID *string
	// 移動ログの更新
	if input.Moved || input.PreviousJoinID == nil {
		if input.PreviousJoinID != nil {
			_, err = tx.Exec(
				`UPDATE pet_group_joins SET left_at = $1, updated_at = $1 WHERE id = $2 AND left_at IS NULL`,
				input.SimulatedAt,
				*input.PreviousJoinID,
			)
			if err != nil {
				return false, err
			}
		}

		joinID := domain.NewUUIDString()
		_, err = tx.Exec(
			`INSERT INTO pet_group_joins (
				id,
				pet_id,
				group_master_id,
				joined_at,
				move_reason,
				created_at,
				updated_at
			) VALUES ($1, $2, $3, $4, $5, $4, $4)`,
			joinID,
			input.PetID,
			input.NextGroupID,
			input.SimulatedAt,
			input.MoveReason,
		)
		if err != nil {
			return false, err
		}
		newJoinID = &joinID
	} else {
		newJoinID = input.PreviousJoinID
	}

	// ペットのステータス更新
	_, err = tx.Exec(
		`UPDATE pets
		SET
			current_group_master_id = $1,
			energy = LEAST(100.0000, GREATEST(0.0000, energy + $2)),
			curiosity = LEAST(100.0000, GREATEST(0.0000, curiosity + $3)),
			sociality = LEAST(100.0000, GREATEST(0.0000, sociality + $4)),
			routine = LEAST(100.0000, GREATEST(0.0000, routine + $5)),
			updated_at = $6
		WHERE id = $7`,
		input.NextGroupID,
		input.EnergyDelta,
		input.CuriosityDelta,
		input.SocialityDelta,
		input.RoutineDelta,
		input.SimulatedAt,
		input.PetID,
	)
	if err != nil {
		return false, err
	}

	// 取得した経験値を加算
	_, err = tx.Exec(
		`UPDATE pet_experiences
		SET
			total_experience = total_experience + $1,
			updated_at = $2
		WHERE pet_id = $3`,
		input.ExperienceAmount,
		input.SimulatedAt,
		input.PetID,
	)
	if err != nil {
		return false, err
	}

	// 経験値取得履歴を作成
	_, err = tx.Exec(
		`INSERT INTO pet_experiences_events (
			id,
			pet_id,
			experience_delta,
			feed_count,
			occurred_at
		) VALUES ($1, $2, $3, 0, $4)`,
		domain.NewUUIDString(),
		input.PetID,
		input.ExperienceAmount,
		input.SimulatedAt,
	)
	if err != nil {
		return false, err
	}

	// ログを保存
	log := input.Log
	_, err = tx.Exec(
		`INSERT INTO pet_hourly_logs (
			id,
			pet_id,
			group_master_id,
			pet_group_join_id,
			simulated_at,
			stayed,
			move_probability,
			boredom,
			rest_need,
			current_group_fit,
			attachment_to_current_group,
			recent_move_penalty,
			energy_delta_applied,
			curiosity_delta_applied,
			sociality_delta_applied,
			routine_delta_applied,
			interaction_count,
			ambient_event,
			report_material,
			created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19, $20
		)`,
		log.ID(),
		log.PetID(),
		log.GroupMasterID(),
		newJoinID,
		log.SimulatedAt(),
		log.Stayed(),
		log.MoveProbability(),
		log.Boredom(),
		log.RestNeed(),
		log.CurrentGroupFit(),
		log.AttachmentToCurrentGroup(),
		log.RecentMovePenalty(),
		log.EnergyDeltaApplied(),
		log.CuriosityDeltaApplied(),
		log.SocialityDeltaApplied(),
		log.RoutineDeltaApplied(),
		log.InteractionCount(),
		log.AmbientEvent(),
		log.ReportMaterial(),
		log.CreatedAt(),
	)
	if err != nil {
		return false, err
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}

	return true, nil
}

type simulationPetScanner interface {
	Scan(dest ...any) error
}

func scanSimulationPet(scanner simulationPetScanner) (domain.Pet, *string, *time.Time, error) {
	var (
		id                   string
		name                 string
		color                string
		isDeleted            bool
		userID               string
		energy               float64
		curiosity            float64
		sociality            float64
		routine              float64
		currentGroupMasterID sql.NullInt64
		currentStageID       int
		createdAt            sql.NullTime
		updatedAt            sql.NullTime
		currentJoinID        sql.NullString
		joinedAt             sql.NullTime
	)

	if err := scanner.Scan(
		&id,
		&name,
		&color,
		&isDeleted,
		&userID,
		&energy,
		&curiosity,
		&sociality,
		&routine,
		&currentGroupMasterID,
		&currentStageID,
		&createdAt,
		&updatedAt,
		&currentJoinID,
		&joinedAt,
	); err != nil {
		return domain.Pet{}, nil, nil, err
	}

	var groupMasterID *int
	if currentGroupMasterID.Valid {
		value := int(currentGroupMasterID.Int64)
		groupMasterID = &value
	}

	var joinID *string
	if currentJoinID.Valid {
		value := currentJoinID.String
		joinID = &value
	}

	var joinedAtValue *time.Time
	if joinedAt.Valid {
		value := joinedAt.Time
		joinedAtValue = &value
	}

	return domain.NewPet(
		domain.PetID(id),
		name,
		color,
		isDeleted,
		domain.UserID(userID),
		energy,
		curiosity,
		sociality,
		routine,
		groupMasterID,
		currentStageID,
		createdAt.Time,
		updatedAt.Time,
	), joinID, joinedAtValue, nil
}
