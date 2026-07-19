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

// 謚慕ｨｿ菫晏ｭ倥∫ｵ碁ｨ灘､蜉邂励∫ｵ碁ｨ灘､螻･豁ｴ菫晏ｭ倥・ｲ蛹門愛螳壹ｒ蜷御ｸ繝医Λ繝ｳ繧ｶ繧ｯ繧ｷ繝ｧ繝ｳ縺ｧ螳溯｡・
func (r *PostRepository) CreateWithFeedExperience(post domain.Post, experienceAmount int) (domain.Post, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return domain.Post{}, mapPersistenceError(err)
	}
	defer tx.Rollback()

	query := `INSERT INTO posts (id, pet_id, content, content_embedding, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err = tx.Exec(query, post.ID(), post.PetID(), post.Content(), post.ContentEmbedding(), post.CreatedAt())
	if err != nil {
		return domain.Post{}, mapPersistenceError(err)
	}

	petExperienceRepo := NewPetExperienceRepository(r.DB)
	if err := petExperienceRepo.AddFeedExperienceTx(tx, post.PetID(), experienceAmount, post.CreatedAt()); err != nil {
		return domain.Post{}, mapPersistenceError(err)
	}

	sourceID := string(post.ID())
	experienceEvent := domain.NewPetExperienceEvent(
		domain.NewPetExperienceEventID(),
		post.PetID(),
		domain.ExperienceSourceTypeFeed,
		&sourceID,
		experienceAmount,
		0,
		post.CreatedAt(),
		post.CreatedAt(),
	)
	petExperienceEventRepo := NewPetExperienceEventRepository(r.DB)
	if err := petExperienceEventRepo.CreateTx(tx, experienceEvent); err != nil {
		return domain.Post{}, mapPersistenceError(err)
	}

	evolutionRuleRepo := NewEvolutionRuleRepository(r.DB)
	satisfiedRule, err := evolutionRuleRepo.FindSatisfiedAfterFeedTx(tx, post.PetID(), post.CreatedAt())
	if err != nil {
		return domain.Post{}, mapPersistenceError(err)
	}

	if satisfiedRule != nil {
		petEvolutionRepo := NewPetEvolutionRepository(r.DB)
		if err := petEvolutionRepo.ApplySatisfiedRuleTx(tx, post.PetID(), *satisfiedRule, post.CreatedAt()); err != nil {
			return domain.Post{}, mapPersistenceError(err)
		}
	}

	if err := tx.Commit(); err != nil {
		return domain.Post{}, mapPersistenceError(err)
	}

	return post, nil
}

// 謖・ｮ壹・繝・ヨ縺ｮ謚慕ｨｿ荳隕ｧ繧呈眠縺励＞鬆・〒蜿門ｾ・
func (r *PostRepository) FindByPetID(petID domain.PetID) ([]domain.Post, error) {
	query := `SELECT id, content, content_embedding, pet_id, created_at FROM posts WHERE pet_id = $1 ORDER BY created_at DESC`

	var posts []domain.Post
	rows, err := r.DB.Query(query, petID)
	if err != nil {
		return nil, mapPersistenceError(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var content string
		var contentEmbedding sql.NullString
		var petIDStr string
		var createdAt time.Time

		if err := rows.Scan(&id, &content, &contentEmbedding, &petIDStr, &createdAt); err != nil {
			return nil, mapPersistenceError(err)
		}

		var embPtr *string
		if contentEmbedding.Valid {
			embPtr = &contentEmbedding.String
		}

		p := domain.NewPost(domain.PostID(id), content, embPtr, domain.PetID(petIDStr), createdAt)
		posts = append(posts, p)
	}

	if err := rows.Err(); err != nil {
		return nil, mapPersistenceError(err)
	}

	return posts, nil
}
