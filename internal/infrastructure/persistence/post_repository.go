package persistence

import (
	"database/sql"

	"github.com/pasokatazip/backend/internal/domain"
)

type PostRepository struct{
	DB *sql.DB
}

func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{DB:db}
}

func (r *PostRepository) Create(post domain.Post) (domain.Post, error) {
	query := `INSERT INTO posts (id, pet_id, content, content_embedding, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.DB.Exec(query, post.ID(), post.PetID(), post.Content(), post.ContentEmbedding(), post.CreatedAt())
	if err != nil {
		return domain.Post{}, err
	}

	return post, nil
}