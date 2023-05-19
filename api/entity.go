package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type Entity struct {
	ID     string
	Secret string
	Scope  Scope
	Attrs  Attrs
}

// CreateEntity creates a new entity. The secret should contain a hashed string, not clear text.
func (a *DefaultAPI) CreateEntity(ctx context.Context, id string, secret []byte, scope Scope, attrs Attrs) error {
	var err error

	if _, err = uuid.Parse(id); err != nil {
		return ErrInvalidArg{Msg: fmt.Sprintf("invalid id: %s", err.Error())}
	}

	if _, err = bcrypt.Cost(secret); err != nil {
		return ErrInvalidArg{Msg: fmt.Sprintf("invalid secretKey: %s", err.Error())}
	}

	scopeArg := pq.StringArray{}
	for _, s := range scope {
		scopeArg = append(scopeArg, s)
	}

	attrsJSON := []byte("{}")
	if len(attrs) != 0 {
		// NOTE: this code is not covered by unit tests in favor of simplicity
		if attrsJSON, err = json.Marshal(attrs); err != nil {
			return ErrInvalidArg{Msg: fmt.Sprintf("invalid attrs: %s", err.Error())}
		}
	}

	q := `INSERT INTO entity (id, secretKey, scope, attrs) VALUES ($1, $2, $3, $4)`
	_, err = a.db.ExecContext(ctx, q, id, secret, scopeArg, attrsJSON)
	if err != nil {
		return err
	}

	return nil
}

// UpdateEntity updates an existing entity. The secret should contain a hashed string, not clear text. Empty secret
// tells this method that it should not be updated.
func (a *DefaultAPI) UpdateEntity(ctx context.Context, id string, secret []byte, scope Scope, attrs Attrs) error {
	var err error

	if _, err := uuid.Parse(id); err != nil {
		return ErrInvalidArg{Msg: fmt.Sprintf("invalid id: %s", err.Error())}
	}

	scopeArg := pq.StringArray{}
	for _, s := range scope {
		scopeArg = append(scopeArg, s)
	}

	attrsJSON := []byte("{}")
	if len(attrs) != 0 {
		// NOTE: this code is not covered by unit tests in favor of simplicity
		attrsJSON, err = json.Marshal(attrs)
		if err != nil {
			return ErrInvalidArg{Msg: fmt.Sprintf("invalid attrs: %s", err.Error())}
		}
	}

	q := `UPDATE entity SET scope=$1, attrs=$2 WHERE id=$3`
	qArgs := []interface{}{scopeArg, attrsJSON, id}

	if len(secret) != 0 {
		if _, err = bcrypt.Cost(secret); err != nil {
			return ErrInvalidArg{Msg: fmt.Sprintf("invalid secretKey: %s", err.Error())}
		}

		q = `UPDATE entity SET secretKey=$1, scope=$2, attrs=$3 WHERE id=$4`
		qArgs = []interface{}{secret, scopeArg, attrsJSON, id}
	}

	qr, err := a.db.ExecContext(ctx, q, qArgs...)
	if err != nil {
		return err
	}

	if ra, err := qr.RowsAffected(); err != nil {
		return err
	} else if ra == 0 {
		return ErrNotFound
	}

	return nil
}

func (a *DefaultAPI) GetEntity(ctx context.Context, id string) (Entity, error) {
	var (
		secret string
		scope  pq.StringArray
		attrs  map[string]string
	)

	row := a.db.QueryRowContext(ctx, `SELECT secretKey, scope, attrs FROM entity WHERE id=$1`, id)
	if err := row.Scan(&secret, &scope, &attrs); errors.Is(err, sql.ErrNoRows) {
		return Entity{}, ErrNotFound
	} else if err != nil {
		return Entity{}, err
	}

	scopeArg := make([]string, 0)
	for _, s := range scope {
		scopeArg = append(scopeArg, s)
	}

	return Entity{ID: id, Secret: secret, Scope: scopeArg, Attrs: attrs}, nil
}
