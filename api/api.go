package api

import (
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/ashep/a23n/sqldb"
)

type API interface {
	SecretKey() string
	CheckSecret(hashed, secret string) (bool, error)

	CreateEntity(ctx context.Context, id string, secret []byte, scope Scope, attrs Attrs) error
	UpdateEntity(ctx context.Context, id string, secret []byte, scope Scope, attrs Attrs) error
	GetEntity(ctx context.Context, id string) (Entity, error)
	CheckScope(target Scope, required Scope) bool

	CreateToken(subject string, scope []string, ttl time.Duration) Token
	ParseToken(token string) (TokenClaims, error)
}

type DefaultAPI struct {
	db        sqldb.DB
	secretKey string
	now       func() time.Time
}

func NewDefault(db sqldb.DB, secretKey string, now func() time.Time) *DefaultAPI {
	return &DefaultAPI{
		db:        db,
		secretKey: secretKey,
		now:       now,
	}
}

func (a *DefaultAPI) SecretKey() string {
	return a.secretKey
}

func (a *DefaultAPI) CheckSecret(hashed, secret string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(secret))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}
