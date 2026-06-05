package persistence

import (
	"database/sql"
	"time"

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

func (r *UserRepository) FindByEmail(email string) (domain.User, error) {
	query := `SELECT id, email, password, subsc, created_at FROM users WHERE email = $1`
	var (
		id        string
		em        string
		password  string
		subsc     bool
		createdAt time.Time
	)

	row := r.DB.QueryRow(query, email)
	if err := row.Scan(&id, &em, &password, &subsc, &createdAt); err != nil {
		return domain.User{}, err
	}

	user := domain.NewUser(domain.UserID(id), em, password, subsc, createdAt)
	return user, nil
}
