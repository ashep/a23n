package api

import (
	"context"
	"time"

	"github.com/ashep/a23n/sqldb"
)

type API interface {
	CheckSecret(hashed, secret string) (bool, error)

	CreateEntity(ctx context.Context, id string, secret []byte, scope Scope, attrs Attrs) error
	UpdateEntity(ctx context.Context, id string, secret []byte, scope Scope, attrs Attrs) error
	GetEntity(ctx context.Context, id string) (Entity, error)
	CheckScope(target Scope, required Scope) bool

	CreateToken(subject string, scope []string, ttl time.Duration) Token
	GetTokenSignedString(t Token) (string, error)
	ParseToken(token string) (TokenClaims, error)
}

type DefaultAPI struct {
	db     sqldb.DB
	secret string
}

func NewDefault(db sqldb.DB, secret string) *DefaultAPI {
	return &DefaultAPI{
		db:     db,
		secret: secret,
	}
}
