package persistence

import (
	"database/sql"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/timeutil"
)

type PetSimulationRepository struct {
	DB *sql.DB
}

const (
	groupInterestHalfLifeSeconds = 14 * 24 * 60 * 60
	groupInterestMinimumScore    = 0.2
)

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
			AND p.status = 'active'
		ORDER BY p.created_at
	`

	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, mapPersistenceError(err)
	}
	defer rows.Close()

	pets := make([]domain.SimulationPet, 0)
	for rows.Next() {
		pet, currentJoinID, joinedAt, err := scanSimulationPet(rows)
		if err != nil {
			return nil, mapPersistenceError(err)
		}
		pets = append(pets, domain.SimulationPet{
			Pet:           pet,
			CurrentJoinID: currentJoinID,
			JoinedAt:      joinedAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, mapPersistenceError(err)
	}

	return pets, nil
}

func (r *PetSimulationRepository) FindActiveGroupsForSimulation() ([]domain.GroupMaster, error) {
	return NewGroupMasterRepository(r.DB).FindActive()
}

func (r *PetSimulationRepository) PruneExpiredGroupInterestsForSimulation() error {
	_, err := r.DB.Exec(
		`DELETE FROM pet_group_interests pgi
		USING group_masters gm
		WHERE pgi.group_master_id = gm.id
			AND (
				gm.active = FALSE
				OR pgi.last_matched_at < CURRENT_TIMESTAMP - INTERVAL '60 days'
				OR pgi.interest_score * POWER(
					0.5::double precision,
					GREATEST(
						0::double precision,
						EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - pgi.last_matched_at)) / $1
					)
				) < $2
			)`,
		groupInterestHalfLifeSeconds,
		groupInterestMinimumScore,
	)
	return err
}

func (r *PetSimulationRepository) FindGroupInterestsForSimulation() (domain.PetGroupInterests, error) {
	rows, err := r.DB.Query(
		`SELECT
			pgi.pet_id,
			pgi.group_master_id,
			pgi.interest_score * POWER(
				0.5::double precision,
				GREATEST(
					0::double precision,
					EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - pgi.last_matched_at)) / $1
				)
			) AS interest_score
		FROM pet_group_interests pgi
		INNER JOIN pets p ON p.id = pgi.pet_id
		INNER JOIN user_active_pets uap ON uap.pet_id = p.id
		INNER JOIN group_masters gm ON gm.id = pgi.group_master_id
		WHERE p.is_deleted = FALSE
			AND p.status = 'active'
			AND gm.active = TRUE
		ORDER BY pgi.pet_id, pgi.group_master_id`,
		groupInterestHalfLifeSeconds,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	interests := make(domain.PetGroupInterests)
	for rows.Next() {
		var (
			petID         string
			groupMasterID int
			interestScore float64
		)
		if err := rows.Scan(&petID, &groupMasterID, &interestScore); err != nil {
			return nil, err
		}

		petInterests, ok := interests[domain.PetID(petID)]
		if !ok {
			petInterests = make(domain.GroupInterestScores)
			interests[domain.PetID(petID)] = petInterests
		}
		petInterests[domain.GroupMasterID(groupMasterID)] = interestScore
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return interests, nil
}

// 同じ simulated_at・同じ群れにいた別ペットの既存興味を、受け手に小さく伝える候補として返す
// 投稿本文・抽出名詞は参照せず、群れIDと興味スコアだけを扱う
func (r *PetSimulationRepository) FindInterestPropagationCandidates(simulatedAt time.Time) ([]domain.InterestPropagationCandidate, error) {
	rows, err := r.DB.Query(
		`SELECT
			recipient_log.pet_id,
			source_log.pet_id,
			source_log.id,
			recipient_log.group_master_id,
			source_interest.group_master_id,
			source_interest.interest_score * POWER(
				0.5::double precision,
				GREATEST(
					0::double precision,
					EXTRACT(EPOCH FROM ($1 - source_interest.last_matched_at)) / $2
				)
			) AS source_interest_score,
			source_pet.sociality,
			recipient_pet.curiosity
		FROM pet_hourly_logs recipient_log
		INNER JOIN pets recipient_pet ON recipient_pet.id = recipient_log.pet_id
		INNER JOIN user_active_pets recipient_active ON recipient_active.pet_id = recipient_pet.id
		INNER JOIN pet_hourly_logs source_log
			ON source_log.simulated_at = recipient_log.simulated_at
			AND source_log.group_master_id = recipient_log.group_master_id
			AND source_log.pet_id <> recipient_log.pet_id
		INNER JOIN pets source_pet ON source_pet.id = source_log.pet_id
		INNER JOIN user_active_pets source_active ON source_active.pet_id = source_pet.id
		INNER JOIN pet_group_interests source_interest
			ON source_interest.pet_id = source_pet.id
		INNER JOIN group_masters propagated_group ON propagated_group.id = source_interest.group_master_id
		WHERE recipient_log.simulated_at = $1
			AND recipient_pet.is_deleted = FALSE
			AND recipient_pet.status = 'active'
			AND source_pet.is_deleted = FALSE
			AND source_pet.status = 'active'
			AND propagated_group.active = TRUE
			AND source_interest.last_matched_at >= $1 - INTERVAL '60 days'
			AND source_interest.interest_score > 0
		ORDER BY recipient_log.pet_id, source_log.pet_id, source_interest.group_master_id`,
		simulatedAt,
		groupInterestHalfLifeSeconds,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	candidates := make([]domain.InterestPropagationCandidate, 0)
	for rows.Next() {
		var candidate domain.InterestPropagationCandidate
		if err := rows.Scan(
			&candidate.RecipientPetID,
			&candidate.SourcePetID,
			&candidate.SourceHourlyLogID,
			&candidate.ViaGroupMasterID,
			&candidate.PropagatedGroupMasterID,
			&candidate.SourceInterestScore,
			&candidate.SourceSociality,
			&candidate.RecipientCuriosity,
		); err != nil {
			return nil, err
		}
		candidates = append(candidates, candidate)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return candidates, nil
}

// 履歴を先に確定し、初回保存時だけ累積興味へ加算する
func (r *PetSimulationRepository) SaveInterestPropagation(propagation domain.PetInterestPropagation) (bool, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var saved bool
	err = tx.QueryRow(
		`INSERT INTO pet_interest_propagations (
			id,
			recipient_pet_id,
			source_pet_id,
			source_pet_hourly_log_id,
			via_group_master_id,
			propagated_group_master_id,
			amount,
			occurred_at,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
		ON CONFLICT (recipient_pet_id, source_pet_hourly_log_id, propagated_group_master_id)
		DO NOTHING
		RETURNING TRUE`,
		domain.NewUUIDString(),
		propagation.RecipientPetID,
		propagation.SourcePetID,
		propagation.SourceHourlyLogID,
		propagation.ViaGroupMasterID,
		propagation.PropagatedGroupMasterID,
		propagation.Amount,
		propagation.OccurredAt,
	).Scan(&saved)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	_, err = tx.Exec(
		`INSERT INTO pet_group_interests (
			id,
			pet_id,
			group_master_id,
			interest_score,
			last_matched_at,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $5, $5)
		ON CONFLICT (pet_id, group_master_id) DO UPDATE
		SET
			interest_score = pet_group_interests.interest_score * POWER(
				0.5::double precision,
				GREATEST(
					0::double precision,
					EXTRACT(EPOCH FROM (EXCLUDED.last_matched_at - pet_group_interests.last_matched_at)) / $6
				)
			) + EXCLUDED.interest_score,
			-- 過去時刻のシミュレーションを再実行しても、既存の興味時刻を巻き戻さない。
			last_matched_at = GREATEST(pet_group_interests.last_matched_at, EXCLUDED.last_matched_at),
			updated_at = GREATEST(pet_group_interests.updated_at, EXCLUDED.updated_at)`,
		domain.NewUUIDString(),
		propagation.RecipientPetID,
		propagation.PropagatedGroupMasterID,
		propagation.Amount,
		propagation.OccurredAt,
		groupInterestHalfLifeSeconds,
	)
	if err != nil {
		return false, err
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}
	return saved, nil
}

// AppendInterestPropagationReportMaterial は、現在いる群れの文頭を残したまま、
// 興味が伝わった群れへの関心を表す1文に差し替える。送り手の情報や投稿本文は保存しない。
func (r *PetSimulationRepository) AppendInterestPropagationReportMaterial(
	petID domain.PetID,
	simulatedAt time.Time,
	propagatedGroupID domain.GroupMasterID,
) error {
	_, err := r.DB.Exec(
		`UPDATE pet_hourly_logs hourly_log
		SET
			ambient_event = '近くの気配から興味を見つけた',
			-- 通常の滞在・休息文に追記せず、現在いる群れの近くで別の群れに興味を持った出来事として残す。
			report_material = current_group.display_name || 'の近くで、' ||
				propagated_group.display_name || 'の気配が少し気になったようです。'
		FROM group_masters propagated_group, group_masters current_group
		WHERE hourly_log.pet_id = $1
			AND hourly_log.simulated_at = $2
			AND propagated_group.id = $3
			AND current_group.id = hourly_log.group_master_id
			-- 再実行でも同じ文を繰り返し追加しない。
			AND hourly_log.ambient_event IS DISTINCT FROM '近くの気配から興味を見つけた'`,
		petID,
		simulatedAt,
		propagatedGroupID,
	)
	return err
}

func (r *PetSimulationRepository) SaveHourlySimulation(input domain.PetSimulationSaveInput) (bool, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return false, mapPersistenceError(err)
	}
	defer tx.Rollback()

	var exists bool
	if err := tx.QueryRow(
		`SELECT EXISTS (SELECT 1 FROM pet_hourly_logs WHERE pet_id = $1 AND simulated_at = $2)`,
		input.PetID,
		input.SimulatedAt,
	).Scan(&exists); err != nil {
		return false, mapPersistenceError(err)
	}
	if exists {
		return false, nil
	}

	var newJoinID *string
	if input.Moved || input.PreviousJoinID == nil {
		if input.PreviousJoinID != nil {
			_, err = tx.Exec(
				`UPDATE pet_group_joins SET left_at = $1, updated_at = $1 WHERE id = $2 AND left_at IS NULL`,
				input.SimulatedAt,
				*input.PreviousJoinID,
			)
			if err != nil {
				return false, mapPersistenceError(err)
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
			return false, mapPersistenceError(err)
		}
		newJoinID = &joinID
	} else {
		newJoinID = input.PreviousJoinID
	}

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
		return false, mapPersistenceError(err)
	}

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
		return false, mapPersistenceError(err)
	}

	if err := saveSouvenirIfDropped(tx, input, log.ID()); err != nil {
		return false, mapPersistenceError(err)
	}

	if err := tx.Commit(); err != nil {
		return false, mapPersistenceError(err)
	}

	return true, nil
}

func (r *PetSimulationRepository) CreateReportsForSimulation(simulatedAt time.Time) (int, error) {
	result, err := r.DB.Exec(
		`INSERT INTO reports (
			id, user_id, pet_id, hour_slot, gossip, group_master_id, previous_group_master_id,
			moved, behavior_type, behavior_label, interaction_count,
			energy_delta, curiosity_delta, sociality_delta, routine_delta,
			reason_json, rumor, created_at
		)
		SELECT
			md5(hourly_log.pet_id::TEXT || ':' || hourly_log.simulated_at::TEXT)::UUID,
			reporting_pet.user_id,
			hourly_log.pet_id,
			EXTRACT(HOUR FROM hourly_log.simulated_at AT TIME ZONE 'Asia/Tokyo')::INTEGER,
			LEFT(NULLIF(hourly_log.report_material, ''), 255),
			hourly_log.group_master_id,
			NULL,
			NOT hourly_log.stayed,
			CASE WHEN hourly_log.stayed THEN 'stayed' ELSE 'moved' END,
			LEFT(COALESCE(NULLIF(hourly_log.ambient_event, ''), CASE WHEN hourly_log.stayed THEN '群れでゆっくり過ごした' ELSE '別の群れへ移動した' END), 255),
			hourly_log.interaction_count,
			ROUND(hourly_log.energy_delta_applied)::INTEGER,
			ROUND(hourly_log.curiosity_delta_applied)::INTEGER,
			ROUND(hourly_log.sociality_delta_applied)::INTEGER,
			ROUND(hourly_log.routine_delta_applied)::INTEGER,
			jsonb_build_object('source', 'hourly_simulation', 'simulated_at', hourly_log.simulated_at),
			COALESCE(rumor_posts.items, '[]'::jsonb),
			hourly_log.simulated_at
		FROM pet_hourly_logs AS hourly_log
		INNER JOIN pets AS reporting_pet ON reporting_pet.id = hourly_log.pet_id
		LEFT JOIN LATERAL (
			SELECT jsonb_agg(candidate.content ORDER BY candidate.created_at DESC) AS items
			FROM (
				SELECT post.content, post.created_at
				FROM pet_hourly_logs AS nearby_log
				INNER JOIN pets AS nearby_pet ON nearby_pet.id = nearby_log.pet_id
				INNER JOIN posts AS post ON post.pet_id = nearby_pet.id
				WHERE nearby_log.simulated_at = hourly_log.simulated_at
					AND nearby_log.group_master_id = hourly_log.group_master_id
					AND nearby_pet.user_id <> reporting_pet.user_id
					AND post.created_at <= hourly_log.simulated_at
				ORDER BY post.created_at DESC
				LIMIT 2
			) AS candidate
		) AS rumor_posts ON TRUE
		WHERE hourly_log.simulated_at = $1
		ON CONFLICT (pet_id, created_at) DO NOTHING`,
		simulatedAt,
	)
	if err != nil {
		return 0, mapPersistenceError(err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return 0, mapPersistenceError(err)
	}
	return int(count), nil
}

func saveSouvenirIfDropped(tx *sql.Tx, input domain.PetSimulationSaveInput, hourlyLogID domain.PetHourlyLogID) error {
	if !input.SouvenirDrop {
		return nil
	}

	foundOn := input.SimulatedAt.In(timeutil.LocationJST()).Format("2006-01-02")

	var dailyCount int
	if err := tx.QueryRow(
		`SELECT COUNT(*) FROM pet_souvenirs WHERE pet_id = $1 AND found_on = $2`,
		input.PetID,
		foundOn,
	).Scan(&dailyCount); err != nil {
		return mapPersistenceError(err)
	}
	if dailyCount >= 3 {
		return nil
	}

	var souvenirMasterID int
	err := tx.QueryRow(
		`SELECT id
		FROM souvenir_masters
		WHERE group_master_id = $1
			AND active = TRUE
		LIMIT 1`,
		input.NextGroupID,
	).Scan(&souvenirMasterID)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return mapPersistenceError(err)
	}

	_, err = tx.Exec(
		`INSERT INTO pet_souvenirs (
			id,
			pet_id,
			souvenir_master_id,
			pet_hourly_log_id,
			found_at,
			found_on,
			source_group_master_id,
			note,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $5, $5)`,
		domain.NewUUIDString(),
		input.PetID,
		souvenirMasterID,
		hourlyLogID,
		input.SimulatedAt,
		foundOn,
		input.NextGroupID,
		input.SouvenirNote,
	)
	return mapPersistenceError(err)
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
		return domain.Pet{}, nil, nil, mapPersistenceError(err)
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
