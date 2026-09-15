package persistence

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pasokatazip/backend/internal/domain"
)

type Transaction struct{ DB *sql.DB }

func NewTransaction(db *sql.DB) *Transaction { return &Transaction{DB: db} }

type transactionContextKey struct{}
type transactionScope struct {
	db *sql.DB
	tx *sql.Tx
}

// WithinTransaction はSQLトランザクションを管理する。処理の組み立てはユースケースが担当する。
func (t *Transaction) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	if ctx.Value(transactionContextKey{}) != nil {
		return fmt.Errorf("%w: nested transaction is not supported", domain.ErrInternal)
	}
	tx, err := t.DB.BeginTx(ctx, nil)
	if err != nil {
		return mapPersistenceError(err)
	}
	defer tx.Rollback()
	txCtx := context.WithValue(ctx, transactionContextKey{}, transactionScope{db: t.DB, tx: tx})
	if err := fn(txCtx); err != nil {
		return err
	}
	return mapPersistenceError(tx.Commit())
}

// requireTransaction はトランザクション必須の書き込みが自動コミットで実行されることを防ぐ。
func requireTransaction(ctx context.Context, db *sql.DB) (*sql.Tx, error) {
	scope, ok := ctx.Value(transactionContextKey{}).(transactionScope)
	if !ok || scope.tx == nil || scope.db != db {
		return nil, fmt.Errorf("%w: repository requires a transaction for its database", domain.ErrInternal)
	}
	return scope.tx, nil
}
