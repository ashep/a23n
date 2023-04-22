package api

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/ashep/a23n/sqldb"
)

type API interface {
	CheckSecret(hashed, secret string) (bool, error)

	CreateEntity(ctx context.Context, uuidGen UUIDGenerator, passwdHashGen PasswordHashGenerator, secret string, scope Scope, attrs Attrs) (Entity, error)
	GetEntity(ctx context.Context, id string) (Entity, error)
	UpdateEntity(ctx context.Context, id, secret string, scope []string, attrs map[string]string) error
	CheckScope(target Scope, required Scope) bool

	CreateToken(subject string, scope []string, ttl time.Duration) *jwt.Token
	GetTokenSignedString(t *jwt.Token) (string, error)
	ParseToken(token string) (TokenClaims, error)
}

type UUIDGenerator func() string

type PasswordHashGenerator func(password []byte, cost int) ([]byte, error)

type DefaultAPI struct {
	db     sqldb.DB
	phg    PasswordHashGenerator
	secret string
}

func NewDefault(db sqldb.DB, secret string) *DefaultAPI {
	return &DefaultAPI{
		db:     db,
		secret: secret,
	}
}
