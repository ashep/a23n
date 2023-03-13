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
	Scope  []string
	Note   string
}

func (a *API) CreateEntity(ctx context.Context, secret, note string, scope []string) (string, error) {
	if len(scope) == 0 {
		return "", ErrInvalidArg{Msg: "empty scope"}
	}

	secretHash, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash a secret: %w", err)
	}

	id := uuid.NewString()

	scopeArg := pq.StringArray{}
	for _, s := range scope {
		scopeArg = append(scopeArg, s)
	}

	q := `INSERT INTO entity (id, secret, scope, note) VALUES ($1, $2, $3, $4)`
	_, err = a.db.ExecContext(ctx, q, id, secretHash, scopeArg, note)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (a *API) UpdateEntity(ctx context.Context, id, secret, note string, scope []string) error {
	var err error

	if _, err = uuid.Parse(id); err != nil {
		return ErrInvalidArg{Msg: fmt.Sprintf("invalid id: %s", err.Error())}
	}

	if len(scope) == 0 {
		return ErrInvalidArg{Msg: "empty scope"}
	}

	scopeArg := pq.StringArray{}
	for _, s := range scope {
		scopeArg = append(scopeArg, s)
	}

	var qr sql.Result
	if secret != "" {
		var secretHash []byte
		if secretHash, err = bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost); err != nil {
			return fmt.Errorf("failed to hash a secret: %w", err)
		}
		qr, err = a.db.ExecContext(ctx, `UPDATE entity SET secret=$1, scope=$2, note=$3 WHERE id=$4`,
			secretHash, scopeArg, note, id)
	} else {
		qr, err = a.db.ExecContext(ctx, `UPDATE entity SET scope=$1, note=$2 WHERE id=$3`, scopeArg, note, id)
	}

	if err != nil {
		return err
	}

	if ra, err := qr.RowsAffected(); err != nil {
		return err
	} else if ra == 0 {
		return ErrNotFound
	}

	return err
}

func (a *API) GetEntity(ctx context.Context, id string) (Entity, error) {
	var (
		secret string
		scope  pq.StringArray
		note   string
	)

	row := a.db.QueryRowContext(ctx, `SELECT secret, scope, note FROM entity WHERE id=$1`, id)
	if err := row.Scan(&secret, &scope, &note); errors.Is(err, sql.ErrNoRows) {
		return Entity{}, ErrNotFound
	} else if err != nil {
		return Entity{}, err
	}

	scopeArg := make([]string, 0)
	for _, s := range scope {
		scopeArg = append(scopeArg, s)
	}

	return Entity{ID: id, Secret: secret, Scope: scopeArg, Note: note}, nil
}
