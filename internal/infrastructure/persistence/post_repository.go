package persistence

import (
	"context"
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

// Create はユースケースが開始したトランザクション内で投稿を保存する。
func (r *PostRepository) Create(ctx context.Context, post domain.Post) (domain.Post, error) {
	tx, err := requireTransaction(ctx, r.DB)
	if err != nil {
		return domain.Post{}, err
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO posts (id, pet_id, content, content_embedding, created_at) VALUES ($1, $2, $3, $4, $5)`, post.ID(), post.PetID(), post.Content(), post.ContentEmbedding(), post.CreatedAt())
	if err != nil {
		return domain.Post{}, mapPersistenceError(err)
	}
	return post, nil
}

func (r *PostRepository) FindByPetID(ctx context.Context, petID domain.PetID) ([]domain.Post, error) {
	query := `SELECT id, content, content_embedding, pet_id, created_at FROM posts WHERE pet_id = $1 ORDER BY created_at DESC`

	var posts []domain.Post
	rows, err := r.DB.QueryContext(ctx, query, petID)
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
