package pq

import (
	"database/sql"
	"database/sql/driver"
	"errors"
)

// ErrDriverUnavailable signals that the stub driver is active in offline mode.
var ErrDriverUnavailable = errors.New("pq driver is unavailable in offline environment")

func init() {
	sql.Register("postgres", Driver{})
}

type Driver struct{}

type Conn struct{}

type Tx struct{}

type Stmt struct{}

type Rows struct{}

type Result struct{}

func (Driver) Open(name string) (driver.Conn, error) {
	return Conn{}, ErrDriverUnavailable
}

func (Conn) Prepare(query string) (driver.Stmt, error) { return Stmt{}, errors.New("not implemented") }
func (Conn) Close() error                              { return nil }
func (Conn) Begin() (driver.Tx, error)                 { return Tx{}, errors.New("not implemented") }

func (Tx) Commit() error   { return nil }
func (Tx) Rollback() error { return nil }

func (Stmt) Close() error  { return nil }
func (Stmt) NumInput() int { return -1 }
func (Stmt) Exec(args []driver.Value) (driver.Result, error) {
	return Result{}, errors.New("not implemented")
}
func (Stmt) Query(args []driver.Value) (driver.Rows, error) {
	return Rows{}, errors.New("not implemented")
}

func (Rows) Columns() []string { return nil }
func (Rows) Close() error      { return nil }
func (Rows) Next(dest []driver.Value) error {
	return errors.New("not implemented")
}

func (Result) LastInsertId() (int64, error) { return 0, errors.New("not implemented") }
func (Result) RowsAffected() (int64, error) { return 0, errors.New("not implemented") }
