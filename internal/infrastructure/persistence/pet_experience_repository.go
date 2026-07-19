package persistence

import (
	"database/sql"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type PetExperienceRepository struct {
	DB *sql.DB
}

func NewPetExperienceRepository(db *sql.DB) *PetExperienceRepository {
	return &PetExperienceRepository{DB: db}
}

// 繝壹ャ繝医・邨碁ｨ灘､髮・ｨ医Ξ繧ｳ繝ｼ繝峨ｒ譁ｰ隕丈ｽ懈・
func (r *PetExperienceRepository) Create(petExperience domain.PetExperience) (domain.PetExperience, error) {
	_, err := r.DB.Exec(
		`INSERT INTO pet_experiences (
			id,
			pet_id,
			total_experience,
			feed_count,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)`,
		petExperience.ID(),
		petExperience.PetID(),
		petExperience.TotalExperience(),
		petExperience.FeedCount(),
		petExperience.CreatedAt(),
		petExperience.UpdatedAt(),
	)
	if err != nil {
		return domain.PetExperience{}, mapPersistenceError(err)
	}

	return petExperience, nil
}

// 謖・ｮ壹・繝・ヨ縺ｮ邨碁ｨ灘､髮・ｨ医ｒ蜿門ｾ・
func (r *PetExperienceRepository) FindByPetID(petID domain.PetID) (domain.PetExperience, error) {
	row := r.DB.QueryRow(
		`SELECT
			id,
			pet_id,
			total_experience,
			feed_count,
			created_at,
			updated_at
		FROM pet_experiences
		WHERE pet_id = $1`,
		petID,
	)

	return scanPetExperience(row)
}

// 邨碁ｨ灘､髮・ｨ医・邏ｯ險育ｵ碁ｨ灘､縺ｨfeed蝗樊焚繧呈峩譁ｰ
func (r *PetExperienceRepository) Update(petExperience domain.PetExperience) (domain.PetExperience, error) {
	_, err := r.DB.Exec(
		`UPDATE pet_experiences
		SET
			total_experience = $1,
			feed_count = $2,
			updated_at = $3
		WHERE id = $4`,
		petExperience.TotalExperience(),
		petExperience.FeedCount(),
		petExperience.UpdatedAt(),
		petExperience.ID(),
	)
	if err != nil {
		return domain.PetExperience{}, mapPersistenceError(err)
	}

	return petExperience, nil
}

// 謚慕ｨｿ繧帝｣溘∋縺溷・縺ｮ邨碁ｨ灘､縺ｨfeed蝗樊焚繧貞酔荳繝医Λ繝ｳ繧ｶ繧ｯ繧ｷ繝ｧ繝ｳ蜀・〒蜉邂・
func (r *PetExperienceRepository) AddFeedExperienceTx(tx *sql.Tx, petID domain.PetID, amount int, occurredAt time.Time) error {
	_, err := tx.Exec(
		`INSERT INTO pet_experiences (
			id,
			pet_id,
			total_experience,
			feed_count,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, 1, $4, $4)
		ON CONFLICT (pet_id) DO UPDATE
		SET
			total_experience = pet_experiences.total_experience + EXCLUDED.total_experience,
			feed_count = pet_experiences.feed_count + 1,
			updated_at = EXCLUDED.updated_at`,
		domain.NewPetExperienceID(),
		petID,
		amount,
		occurredAt,
	)
	return mapPersistenceError(err)
}

type petExperienceScanner interface {
	Scan(dest ...any) error
}

// SQL縺ｮ蜿門ｾ礼ｵ先棡繧単etExperience domain縺ｸ螟画鋤
func scanPetExperience(scanner petExperienceScanner) (domain.PetExperience, error) {
	var (
		id              string
		petID           string
		totalExperience int64
		feedCount       int
		createdAt       time.Time
		updatedAt       time.Time
	)

	if err := scanner.Scan(
		&id,
		&petID,
		&totalExperience,
		&feedCount,
		&createdAt,
		&updatedAt,
	); err != nil {
		return domain.PetExperience{}, mapPersistenceError(err)
	}

	return domain.NewPetExperience(
		domain.PetExperienceID(id),
		domain.PetID(petID),
		totalExperience,
		feedCount,
		createdAt,
		updatedAt,
	), nil
}
