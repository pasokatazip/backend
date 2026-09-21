package persistence

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/pasokatazip/backend/internal/domain"
)

type postTransactionDriver struct{}

func (postTransactionDriver) Open(string) (driver.Conn, error) {
	return nil, errors.New("use connector")
}

type postTransactionConnector struct{ conn *postTransactionConn }

func (c postTransactionConnector) Connect(context.Context) (driver.Conn, error) { return c.conn, nil }
func (postTransactionConnector) Driver() driver.Driver                          { return postTransactionDriver{} }

type postTransactionConn struct {
	driver.Conn
	beginErr, commitErr error
	commits, rollbacks  int
}

func (c *postTransactionConn) Close() error { return nil }
func (c *postTransactionConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	if c.beginErr != nil {
		return nil, c.beginErr
	}
	return c, nil
}
func (c *postTransactionConn) Commit() error   { c.commits++; return c.commitErr }
func (c *postTransactionConn) Rollback() error { c.rollbacks++; return nil }

func TestTransaction(t *testing.T) {
	failure := errors.New("transaction failed")
	for _, mode := range []string{"success", "begin", "callback", "commit", "panic"} {
		t.Run(mode, func(t *testing.T) {
			conn := &postTransactionConn{}
			if mode == "begin" {
				conn.beginErr = failure
			}
			if mode == "commit" {
				conn.commitErr = failure
			}
			db := sql.OpenDB(postTransactionConnector{conn})
			defer db.Close()
			called := false
			var err error
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				err = NewTransaction(db).WithinTransaction(context.Background(), func(txCtx context.Context) error {
					called = true
					tx, txErr := requireTransaction(txCtx, db)
					if txErr != nil || tx == nil {
						t.Fatalf("transaction unavailable: %v", txErr)
					}
					if _, err := requireTransaction(context.Background(), db); !errors.Is(err, domain.ErrInternal) {
						t.Fatalf("missing transaction error=%v", err)
					}
					otherDB := sql.OpenDB(postTransactionConnector{conn: &postTransactionConn{}})
					defer otherDB.Close()
					if _, err := requireTransaction(txCtx, otherDB); !errors.Is(err, domain.ErrInternal) {
						t.Fatalf("wrong database error=%v", err)
					}
					if err := NewTransaction(db).WithinTransaction(txCtx, func(context.Context) error { t.Fatal("nested callback must not run"); return nil }); !errors.Is(err, domain.ErrInternal) {
						t.Fatalf("nested transaction error=%v", err)
					}

					if mode == "callback" {
						return failure
					}
					if mode == "panic" {
						panic(failure)
					}
					return nil
				})
			}()
			if called != (mode != "begin") {
				t.Fatalf("callback called=%v", called)
			}
			if mode == "panic" {
				if recovered != failure {
					t.Fatalf("panic=%v", recovered)
				}
			} else if mode == "success" {
				if err != nil {
					t.Fatal(err)
				}
			} else if mode == "begin" || mode == "commit" {
				if !errors.Is(err, domain.ErrInternal) {
					t.Fatalf("error=%v want mapped internal error", err)
				}
			} else if !errors.Is(err, failure) {
				t.Fatalf("error=%v want=%v", err, failure)
			}
			wantCommits, wantRollbacks := 0, 0
			if mode == "success" || mode == "commit" {
				wantCommits = 1
			}
			if mode == "callback" || mode == "panic" {
				wantRollbacks = 1
			}
			if conn.commits != wantCommits || conn.rollbacks != wantRollbacks {
				t.Fatalf("commits=%d rollbacks=%d want=%d/%d", conn.commits, conn.rollbacks, wantCommits, wantRollbacks)
			}
		})
	}
}
