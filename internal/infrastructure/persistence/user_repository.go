package persistence

import (
	"database/sql"

	"github.com/pasokatazip/backend/internal/domain"
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) Create(user domain.User) (domain.User, error) {
	query := `INSERT INTO users (id, email, password, subsc, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.DB.Exec(query, user.ID(), user.Email(), user.Password(), user.Subsc(), user.CreatedAt())
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}
