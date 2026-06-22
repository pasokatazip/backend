package persistence

import (
	"database/sql"
	"github.com/google/uuid"
	"github.com/pasokatazip/backend/internal/domain"
)

type PetRepository struct {
	DB *sql.DB
}

func NewPetRepository(db *sql.DB) *PetRepository {
	return &PetRepository{DB: db}
}

func (r *PetRepository) Create(pet domain.Pet) (domain.Pet, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return domain.Pet{}, err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO pets (
			id,
			name,
			is_deleted,
			user_id,
			energy,
			curiosity,
			sociality,
			routine,
			current_group_master_id,
			current_stage_id,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err = tx.Exec(
		query,
		pet.ID(),
		pet.Name(),
		pet.IsDeleted(),
		pet.UserID(),
		pet.Energy(),
		pet.Curiosity(),
		pet.Sociality(),
		pet.Routine(),
		pet.CurrentGroupMasterID(),
		pet.CurrentStageID(),
		pet.CreatedAt(),
		pet.UpdatedAt(),
	)
	if err != nil {
		return domain.Pet{}, err
	}

	_, err = tx.Exec(
		`INSERT INTO user_active_pets (user_id, pet_id, assigned_at) VALUES ($1, $2, $3)`,
		pet.UserID(),
		pet.ID(),
		pet.CreatedAt(),
	)
	if err != nil {
		return domain.Pet{}, err
	}

	_, err = tx.Exec(
		`INSERT INTO pet_experiences (id, pet_id, total_experience, feed_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		uuid.New().String(),
		pet.ID(),
		0,
		0,
		pet.CreatedAt(),
		pet.UpdatedAt(),
	)
	if err != nil {
		return domain.Pet{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Pet{}, err
	}

	return pet, nil
}

func (r *PetRepository) FindByID(id domain.PetID) (domain.Pet, error) {
	query := `
		SELECT
			id,
			name,
			is_deleted,
			user_id,
			energy,
			curiosity,
			sociality,
			routine,
			current_group_master_id,
			current_stage_id,
			created_at,
			updated_at
		FROM pets
		WHERE id = $1
	`

	return r.scanPet(r.DB.QueryRow(query, id))
}

func (r *PetRepository) FindActiveByUserID(userID domain.UserID) (domain.Pet, error) {
	query := `
		SELECT
			p.id,
			p.name,
			p.is_deleted,
			p.user_id,
			p.energy,
			p.curiosity,
			p.sociality,
			p.routine,
			p.current_group_master_id,
			p.current_stage_id,
			p.created_at,
			p.updated_at
		FROM pets p
		INNER JOIN user_active_pets uap ON uap.pet_id = p.id
		WHERE uap.user_id = $1
		AND p.is_deleted = false
		ORDER BY uap.assigned_at DESC
		LIMIT 1
	`

	return r.scanPet(r.DB.QueryRow(query, userID))
}

func (r *PetRepository) scanPet(row *sql.Row) (domain.Pet, error) {
	var (
		id                   string
		name                 string
		isDeleted            bool
		userID               string
		energy               int
		curiosity            int
		sociality            int
		routine              int
		currentGroupMasterID sql.NullInt64
		currentStageID       int
		createdAt            sql.NullTime
		updatedAt            sql.NullTime
	)

	if err := row.Scan(
		&id,
		&name,
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
	); err != nil {
		return domain.Pet{}, err
	}

	var groupMasterID *int
	if currentGroupMasterID.Valid {
		value := int(currentGroupMasterID.Int64)
		groupMasterID = &value
	}

	return domain.NewPet(
		domain.PetID(id),
		name,
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
	), nil
}
