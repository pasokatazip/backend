package persistence

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pasokatazip/backend/internal/domain"
)

// mapPersistenceError prevents database-specific errors from crossing the
// infrastructure boundary while retaining their details for server logs.
func mapPersistenceError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, domain.ErrValidation) ||
		errors.Is(err, domain.ErrNotFound) ||
		errors.Is(err, domain.ErrAlreadyExists) {
		return err
	}
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return domain.ErrAlreadyExists
	}

	return fmt.Errorf("%w: %v", domain.ErrInternal, err)
}
