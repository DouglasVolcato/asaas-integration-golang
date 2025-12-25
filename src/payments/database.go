package payments

import (
	"context"
	"database/sql"
)

// RowScanner is the minimal interface implemented by *sql.Row and custom rows in tests.
type RowScanner interface {
	Scan(dest ...any) error
}

// DBTransaction represents a unit of work capable of queries and executions.
type DBTransaction interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) RowScanner
	Commit() error
	Rollback() error
}

// Database captures the operations required by this package. Implementations can wrap *sql.DB or provide fakes for tests.
type Database interface {
	PingContext(ctx context.Context) error
	BeginTx(ctx context.Context, opts *sql.TxOptions) (DBTransaction, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// sqlDatabase adapts *sql.DB to the Database interface.
type sqlDatabase struct {
	db *sql.DB
}

// NewSQLDatabase wraps a standard *sql.DB for use with the Client.
func NewSQLDatabase(db *sql.DB) Database {
	return &sqlDatabase{db: db}
}

func (s *sqlDatabase) PingContext(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *sqlDatabase) BeginTx(ctx context.Context, opts *sql.TxOptions) (DBTransaction, error) {
	tx, err := s.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &sqlTransaction{tx: tx}, nil
}

func (s *sqlDatabase) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return s.db.ExecContext(ctx, query, args...)
}

// sqlTransaction adapts *sql.Tx to DBTransaction.
type sqlTransaction struct {
	tx *sql.Tx
}

func (s *sqlTransaction) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return s.tx.ExecContext(ctx, query, args...)
}

func (s *sqlTransaction) QueryRowContext(ctx context.Context, query string, args ...any) RowScanner {
	return s.tx.QueryRowContext(ctx, query, args...)
}

func (s *sqlTransaction) Commit() error {
	return s.tx.Commit()
}

func (s *sqlTransaction) Rollback() error {
	return s.tx.Rollback()
}
