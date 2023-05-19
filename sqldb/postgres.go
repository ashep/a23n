package sqldb

import (
	"context"
	"database/sql"
)

type Postgres struct {
	db *sql.DB
}

func NewPostgres(dsn string) (DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	return &Postgres{db: db}, nil
}

func (p *Postgres) DB() *sql.DB {
	return p.db
}

func (p *Postgres) PingContext(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

func (p *Postgres) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return p.db.ExecContext(ctx, query, args...)
}

func (p *Postgres) QueryRowContext(ctx context.Context, query string, args ...interface{}) Row {
	return p.db.QueryRowContext(ctx, query, args)
}
