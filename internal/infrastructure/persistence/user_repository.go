package persistence

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pasokatazip/backend/internal/domain"
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) Create(user domain.User) (domain.User, error) {
	query := `
		INSERT INTO users (
			id, email, password, subsc, fincode_customer_id, fincode_subscription_id, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.DB.Exec(
		query,
		user.ID(),
		user.Email(),
		user.Password(),
		user.Subsc(),
		user.FincodeCustomerID(),
		user.FincodeSubscriptionID(),
		user.CreatedAt(),
	)
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (r *UserRepository) FindByEmail(email string) (domain.User, error) {
	query := `
		SELECT id, email, password, subsc, fincode_customer_id, fincode_subscription_id, created_at
		FROM users
		WHERE email = $1
	`
	return scanUser(r.DB.QueryRow(query, email))
}

func (r *UserRepository) FindByID(id domain.UserID) (domain.User, error) {
	query := `
		SELECT id, email, password, subsc, fincode_customer_id, fincode_subscription_id, created_at
		FROM users
		WHERE id = $1
	`
	return scanUser(r.DB.QueryRow(query, string(id)))
}

func (r *UserRepository) FindByFincodeCustomerID(customerID string) (domain.User, error) {
	query := `
		SELECT id, email, password, subsc, fincode_customer_id, fincode_subscription_id, created_at
		FROM users
		WHERE fincode_customer_id = $1
	`
	return scanUser(r.DB.QueryRow(query, customerID))
}

func (r *UserRepository) FindByFincodeSubscriptionID(subscriptionID string) (domain.User, error) {
	query := `
		SELECT id, email, password, subsc, fincode_customer_id, fincode_subscription_id, created_at
		FROM users
		WHERE fincode_subscription_id = $1
	`
	return scanUser(r.DB.QueryRow(query, subscriptionID))
}

func (r *UserRepository) UpdateFincodeCustomerID(id domain.UserID, customerID string) error {
	result, err := r.DB.Exec(`
		UPDATE users
		SET fincode_customer_id = $1
		WHERE id = $2
	`, customerID, id)
	if err != nil {
		return err
	}
	return requireUpdatedRow(result)
}

func (r *UserRepository) UpdateFincodeSubscription(
	id domain.UserID,
	subscriptionID string,
	subsc bool,
) error {
	result, err := r.DB.Exec(`
		UPDATE users
		SET
			fincode_subscription_id = $1,
			subsc = $2
		WHERE id = $3
	`, subscriptionID, subsc, id)
	if err != nil {
		return err
	}
	return requireUpdatedRow(result)
}

func (r *UserRepository) UpdateSubscriptionStatus(id domain.UserID, subsc bool) error {
	result, err := r.DB.Exec(`
		UPDATE users
		SET subsc = $1
		WHERE id = $2
	`, subsc, id)
	if err != nil {
		return err
	}
	return requireUpdatedRow(result)
}

func (r *UserRepository) UpdateEmail(id domain.UserID, email string) error {
	result, err := r.DB.Exec(`UPDATE users SET email = $1 WHERE id = $2`, email, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrAlreadyExists
		}
		return err
	}
	return requireUpdatedRow(result)
}

func (r *UserRepository) UpdatePassword(id domain.UserID, password string) error {
	result, err := r.DB.Exec(`UPDATE users SET password = $1 WHERE id = $2`, password, id)
	if err != nil {
		return err
	}
	return requireUpdatedRow(result)
}

type userRowScanner interface {
	Scan(dest ...any) error
}

func scanUser(row userRowScanner) (domain.User, error) {
	var (
		id                    string
		email                 string
		password              string
		subsc                 bool
		fincodeCustomerID     sql.NullString
		fincodeSubscriptionID sql.NullString
		createdAt             time.Time
	)

	if err := row.Scan(
		&id,
		&email,
		&password,
		&subsc,
		&fincodeCustomerID,
		&fincodeSubscriptionID,
		&createdAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, err
	}

	return domain.NewUser(
		domain.UserID(id),
		email,
		password,
		subsc,
		nullableString(fincodeCustomerID),
		nullableString(fincodeSubscriptionID),
		createdAt,
	), nil
}

func requireUpdatedRow(result sql.Result) error {
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func nullableString(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}
