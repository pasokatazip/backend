package persistence

import (
	"database/sql"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type GroupMasterRepository struct {
	DB *sql.DB
}

func NewGroupMasterRepository(db *sql.DB) *GroupMasterRepository {
	return &GroupMasterRepository{DB: db}
}

func (r *GroupMasterRepository) FindActive() ([]domain.GroupMaster, error) {
	query := `
		SELECT
			id,
			group_key,
			display_name,
			category,
			min_pet_count,
			energy_delta,
			curiosity_delta,
			sociality_delta,
			routine_delta,
			active,
			created_at
		FROM group_masters
		WHERE active = TRUE
		ORDER BY id
	`

	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, mapPersistenceError(err)
	}
	defer rows.Close()

	groups := make([]domain.GroupMaster, 0)
	for rows.Next() {
		group, err := scanGroupMaster(rows)
		if err != nil {
			return nil, mapPersistenceError(err)
		}
		groups = append(groups, group)
	}

	if err := rows.Err(); err != nil {
		return nil, mapPersistenceError(err)
	}

	return groups, nil
}

func (r *GroupMasterRepository) FindByID(id domain.GroupMasterID) (domain.GroupMaster, error) {
	query := `
		SELECT
			id,
			group_key,
			display_name,
			category,
			min_pet_count,
			energy_delta,
			curiosity_delta,
			sociality_delta,
			routine_delta,
			active,
			created_at
		FROM group_masters
		WHERE id = $1
	`

	row := r.DB.QueryRow(query, int(id))
	return scanGroupMaster(row)
}

func (r *GroupMasterRepository) FindByGroupKey(groupKey string) (domain.GroupMaster, error) {
	query := `
		SELECT
			id,
			group_key,
			display_name,
			category,
			min_pet_count,
			energy_delta,
			curiosity_delta,
			sociality_delta,
			routine_delta,
			active,
			created_at
		FROM group_masters
		WHERE group_key = $1
	`

	row := r.DB.QueryRow(query, groupKey)
	return scanGroupMaster(row)
}

type groupMasterScanner interface {
	Scan(dest ...any) error
}

func scanGroupMaster(scanner groupMasterScanner) (domain.GroupMaster, error) {
	var (
		id             int
		groupKey       string
		displayName    string
		category       sql.NullString
		minPetCount    int
		energyDelta    float64
		curiosityDelta float64
		socialityDelta float64
		routineDelta   float64
		active         bool
		createdAt      time.Time
	)

	if err := scanner.Scan(
		&id,
		&groupKey,
		&displayName,
		&category,
		&minPetCount,
		&energyDelta,
		&curiosityDelta,
		&socialityDelta,
		&routineDelta,
		&active,
		&createdAt,
	); err != nil {
		return domain.GroupMaster{}, mapPersistenceError(err)
	}

	var categoryValue *string
	if category.Valid {
		categoryValue = &category.String
	}

	return domain.NewGroupMaster(
		domain.GroupMasterID(id),
		groupKey,
		displayName,
		categoryValue,
		minPetCount,
		energyDelta,
		curiosityDelta,
		socialityDelta,
		routineDelta,
		active,
		createdAt,
	), nil
}
