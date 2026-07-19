package database

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pasokatazip/backend/internal/domain"
)

func NewPostgresDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("%w: open postgres: %v", domain.ErrInternal, err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("%w: ping postgres: %v", domain.ErrInternal, err)
	}

	return db, nil
}
