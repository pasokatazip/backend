package persistence

import (
	"database/sql"
	"errors"

	"github.com/pasokatazip/backend/internal/domain"
)

type PetSouvenirRepository struct {
	DB *sql.DB
}

func NewPetSouvenirRepository(db *sql.DB) *PetSouvenirRepository {
	return &PetSouvenirRepository{DB: db}
}

func (r *PetSouvenirRepository) FindLatestByActivePetUserID(
	userID domain.UserID,
) (*domain.PetSouvenir, error) {
	row := r.DB.QueryRow(
		`SELECT
			latest.id,
			latest.display_name,
			latest.image_url,
			latest.found_at,
			latest.reported
		FROM user_active_pets uap
		INNER JOIN pets p
			ON p.id = uap.pet_id
			AND p.user_id = uap.user_id
			AND p.is_deleted = FALSE
			AND p.status = 'active'
		LEFT JOIN LATERAL (
			SELECT
				ps.id,
				sm.display_name,
				COALESCE(sm.image_url, '') AS image_url,
				ps.found_at,
				(ps.report_id IS NOT NULL OR ps.reported_at IS NOT NULL) AS reported
			FROM pet_souvenirs ps
			INNER JOIN souvenir_masters sm ON sm.id = ps.souvenir_master_id
			WHERE ps.pet_id = p.id
			ORDER BY ps.found_at DESC, ps.created_at DESC, ps.id DESC
			LIMIT 1
		) latest ON TRUE
		WHERE uap.user_id = $1`,
		userID,
	)

	var (
		id          sql.NullString
		displayName sql.NullString
		imageURL    sql.NullString
		foundAt     sql.NullTime
		reported    sql.NullBool
	)
	if err := row.Scan(&id, &displayName, &imageURL, &foundAt, &reported); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, mapPersistenceError(err)
	}

	if !id.Valid {
		return nil, nil
	}

	souvenir := domain.NewPetSouvenir(
		id.String,
		displayName.String,
		imageURL.String,
		foundAt.Time,
		reported.Bool,
	)
	return &souvenir, nil
}
