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

func (r *PostRepository) Create(post domain.Post) (domain.Post, error) {
	query := `INSERT INTO posts (id, pet_id, content, content_embedding, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.DB.Exec(query, post.ID(), post.PetID(), post.Content(), post.ContentEmbedding(), post.CreatedAt())
	if err != nil {
		return domain.Post{}, err
	}

	return post, nil
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
