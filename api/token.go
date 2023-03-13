package api

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenClaims struct {
	jwt.RegisteredClaims
}

func (a *API) CreateToken(e Entity) *jwt.Token {
	n := jwt.NewNumericDate(time.Now())
	exp := jwt.NewNumericDate(n.Add(time.Duration(a.tokenTTL) * time.Second))

	return jwt.NewWithClaims(jwt.SigningMethodHS256, TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   e.ID,
			IssuedAt:  n,
			NotBefore: n,
			ExpiresAt: exp,
		},
	})
}

func (a *API) GetTokenSignedString(t *jwt.Token) (string, error) {
	return t.SignedString([]byte(a.secret))
}

func (a *API) GetEntityByTokenString(ctx context.Context, token string) (Entity, error) {
	clm := TokenClaims{}
	_, err := jwt.ParseWithClaims(token, &clm, func(token *jwt.Token) (interface{}, error) {
		return []byte(a.secret), nil
	})
	if err != nil {
		return Entity{}, err
	}

	e, err := a.GetEntity(ctx, clm.Subject)
	if err != nil {
		return Entity{}, err
	}

	return e, nil
}
