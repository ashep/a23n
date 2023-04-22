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
	Scope  Scope
	Attrs  Attrs
}

func (a *DefaultAPI) CreateEntity(
	ctx context.Context,
	uuidGen UUIDGenerator,
	passwdHashGen PasswordHashGenerator,
	secret string,
	scope Scope,
	attrs Attrs,
) (Entity, error) {
	if len(scope) == 0 {
		return Entity{}, ErrInvalidArg{Msg: "empty scope"}
	}

	secretHash, err := passwdHashGen([]byte(secret), bcrypt.DefaultCost)
	if err != nil {
		return Entity{}, fmt.Errorf("failed to hash a secret: %w", err)
	}

	id := uuidGen()

	scopeArg := pq.StringArray{}
	for _, s := range scope {
		scopeArg = append(scopeArg, s)
	}

	q := `INSERT INTO entity (id, secret, scope, attrs) VALUES ($1, $2, $3, $4)`
	_, err = a.db.ExecContext(ctx, q, id, secretHash, scopeArg, attrs)
	if err != nil {
		return Entity{}, err
	}

	e := Entity{
		ID:     id,
		Secret: string(secretHash),
		Scope:  scope,
		Attrs:  attrs,
	}

	return e, nil
}

func (a *DefaultAPI) UpdateEntity(ctx context.Context, id, secret string, scope []string, attrs map[string]string) error {
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
		qr, err = a.db.ExecContext(ctx, `UPDATE entity SET secret=$1, scope=$2, attrs=$3 WHERE id=$4`,
			secretHash, scopeArg, attrs, id)
	} else {
		qr, err = a.db.ExecContext(ctx, `UPDATE entity SET scope=$1, attrs=$2 WHERE id=$3`, scopeArg, attrs, id)
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

func (a *DefaultAPI) GetEntity(ctx context.Context, id string) (Entity, error) {
	var (
		secret string
		scope  pq.StringArray
		attrs  map[string]string
	)

	row := a.db.QueryRowContext(ctx, `SELECT secret, scope, attrs FROM entity WHERE id=$1`, id)
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
