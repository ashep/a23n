package sqldb

import (
	"context"
	"database/sql"
)

type Row interface {
	Scan(args ...interface{}) error
	Err() error
}

type DB interface {
	DB() *sql.DB
	PingContext(ctx context.Context) error
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}
