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

// ペットの経験値集計レコードを新規作成
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
		return domain.PetExperience{}, err
	}

	return petExperience, nil
}

// 指定ペットの経験値集計を取得
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

// 経験値集計の累計経験値とfeed回数を更新
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
		return domain.PetExperience{}, err
	}

	return petExperience, nil
}

// 投稿を食べた分の経験値とfeed回数を同一トランザクション内で加算
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
	return err
}

type petExperienceScanner interface {
	Scan(dest ...any) error
}

// SQLの取得結果をPetExperience domainへ変換
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
		return domain.PetExperience{}, err
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
