package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type Entity struct {
	ID     string
	Secret string
	Attrs  []string
	Note   string
}

func (a *API) GetEntity(ctx context.Context, id string) (Entity, error) {
	var (
		secret string
		attrs  pq.StringArray
		note   string
	)

	row := a.db.QueryRowContext(ctx, `SELECT secret, attrs, note FROM entity WHERE id=$1`, id)
	if err := row.Scan(&secret, &attrs, &note); errors.Is(err, sql.ErrNoRows) {
		return Entity{}, ErrNotFound
	} else if err != nil {
		return Entity{}, err
	}

	attrsArg := make([]string, 0)
	for _, attr := range attrs {
		attrsArg = append(attrsArg, attr)
	}

	return Entity{ID: id, Secret: secret, Attrs: attrsArg, Note: note}, nil
}

func (a *API) CreateEntity(
	ctx context.Context,
	adminSecret string,
	entitySecret string,
	attrs []string,
	note string,
) (Entity, error) {
	if adminSecret != a.secret {
		return Entity{}, ErrUnauthorized
	}

	if len(attrs) == 0 {
		return Entity{}, ErrInvalidArg{Msg: "empty attributes set"}
	}

	sec, err := bcrypt.GenerateFromPassword([]byte(entitySecret), bcrypt.DefaultCost)
	if err != nil {
		return Entity{}, fmt.Errorf("failed to hash a secret: %w", err)
	}

	tx, err := a.db.Begin()
	if err != nil {
		return Entity{}, err
	}

	id := uuid.NewString()

	attrsArg := pq.StringArray{}
	for _, attr := range attrs {
		attrsArg = append(attrsArg, attr)
	}

	q := `INSERT INTO entity (id, secret, attrs, note) VALUES ($1, $2, $3, $4)`
	_, err = tx.ExecContext(ctx, q, id, sec, attrsArg, note)
	if err != nil {
		return Entity{}, err
	}

	if err := tx.Commit(); err != nil {
		return Entity{}, err
	}

	return a.GetEntity(ctx, id)
}
