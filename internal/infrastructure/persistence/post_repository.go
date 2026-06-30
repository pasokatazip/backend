package persistence

import (
	"database/sql"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type PostRepository struct {
	DB *sql.DB
}

func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{DB: db}
}

func (r *PostRepository) CreateWithFeedExperience(post domain.Post, experienceAmount int) (domain.Post, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return domain.Post{}, err
	}
	defer tx.Rollback()

	query := `INSERT INTO posts (id, pet_id, content, content_embedding, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err = tx.Exec(query, post.ID(), post.PetID(), post.Content(), post.ContentEmbedding(), post.CreatedAt())
	if err != nil {
		return domain.Post{}, err
	}

	_, err = tx.Exec(
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
		domain.NewUUIDString(),
		post.PetID(),
		experienceAmount,
		post.CreatedAt(),
	)
	if err != nil {
		return domain.Post{}, err
	}

	_, err = tx.Exec(
		`INSERT INTO pet_experience_events (
			id,
			pet_id,
			source_type,
			source_id,
			amount,
			capped_amount,
			experience_date,
			created_at
		) VALUES ($1, $2, $3, $4, $5, 0, $6, $7)`,
		domain.NewUUIDString(),
		post.PetID(),
		"feed",
		post.ID(),
		experienceAmount,
		post.CreatedAt().Format("2006-01-02"),
		post.CreatedAt(),
	)
	if err != nil {
		return domain.Post{}, err
	}

	if err := r.applyEvolutionAfterFeed(tx, post); err != nil {
		return domain.Post{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Post{}, err
	}

	return post, nil
}

func (r *PostRepository) applyEvolutionAfterFeed(tx *sql.Tx, post domain.Post) error {
	_, err := tx.Exec(
		`WITH pet_snapshot AS (
			SELECT
				p.id,
				p.current_stage_id,
				p.created_at,
				p.energy,
				p.curiosity,
				p.sociality,
				p.routine,
				pe.total_experience,
				pe.feed_count,
				COALESCE(
					(SELECT MAX(evolved_at) FROM pet_evolutions WHERE pet_id = p.id),
					p.created_at
				) AS last_evolved_at
			FROM pets p
			INNER JOIN pet_experiences pe ON pe.pet_id = p.id
			WHERE p.id = $1
			FOR UPDATE
		),
		candidate_rule AS (
			SELECT
				er.id AS rule_id,
				er.to_stage_id,
				ps.energy,
				ps.curiosity,
				ps.sociality,
				ps.routine
			FROM evolution_rules er
			INNER JOIN pet_snapshot ps ON ps.current_stage_id = er.from_stage_id
			WHERE ps.total_experience >= er.required_experience
			AND ps.feed_count >= er.required_feed_count
			AND DATE_PART('day', $2::timestamptz - ps.last_evolved_at) >= er.required_days_since_last_evolution
			ORDER BY er.required_experience DESC, er.required_feed_count DESC, er.id
			LIMIT 1
		),
		updated_pet AS (
			UPDATE pets p
			SET
				current_stage_id = cr.to_stage_id,
				updated_at = $2
			FROM candidate_rule cr
			WHERE p.id = $1
			RETURNING
				p.id AS pet_id,
				cr.to_stage_id,
				cr.rule_id,
				CASE
					WHEN cr.energy >= cr.curiosity AND cr.energy >= cr.sociality AND cr.energy >= cr.routine THEN 'energy'
					WHEN cr.curiosity >= cr.sociality AND cr.curiosity >= cr.routine THEN 'curiosity'
					WHEN cr.sociality >= cr.routine THEN 'sociality'
					ELSE 'routine'
				END AS primary_status
		)
		INSERT INTO pet_evolutions (
			id,
			pet_id,
			stage_id,
			evolution_rule_id,
			primary_status,
			evolved_at,
			created_at
		)
		SELECT
			$3,
			pet_id,
			to_stage_id,
			rule_id,
			primary_status,
			$2,
			$2
		FROM updated_pet`,
		post.PetID(),
		post.CreatedAt(),
		domain.NewUUIDString(),
	)
	return err
}

func (r *PostRepository) FindByPetID(petID domain.PetID) ([]domain.Post, error) {
	query := `SELECT id, content, content_embedding, pet_id, created_at FROM posts WHERE pet_id = $1 ORDER BY created_at DESC`

	var posts []domain.Post
	rows, err := r.DB.Query(query, petID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var content string
		var contentEmbedding sql.NullString
		var petIDStr string
		var createdAt time.Time

		if err := rows.Scan(&id, &content, &contentEmbedding, &petIDStr, &createdAt); err != nil {
			return nil, err
		}

		var embPtr *string
		if contentEmbedding.Valid {
			embPtr = &contentEmbedding.String
		}

		p := domain.NewPost(domain.PostID(id), content, embPtr, domain.PetID(petIDStr), createdAt)
		posts = append(posts, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}
