package api

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims interface {
	GetExpirationTime() (*jwt.NumericDate, error)
}

type Token interface {
	Claims() Claims
	SignedString(key interface{}) (string, error)
}

type TokenClaims struct {
	jwt.RegisteredClaims
	Scope []string `json:"scope,omitempty"`
}

type DefaultToken struct {
	t *jwt.Token
}

func (t *DefaultToken) Claims() Claims {
	return t.t.Claims
}

func (t *DefaultToken) SignedString(key interface{}) (string, error) {
	return t.t.SignedString(key)
}

func (a *DefaultAPI) CreateToken(subject string, scope []string, ttl time.Duration) Token {
	n := jwt.NewNumericDate(a.now())

	return &DefaultToken{
		t: jwt.NewWithClaims(jwt.SigningMethodHS256, TokenClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   subject,
				IssuedAt:  n,
				NotBefore: n,
				ExpiresAt: jwt.NewNumericDate(n.Add(ttl)),
			},
			Scope: scope,
		}),
	}
}

func (a *DefaultAPI) ParseToken(token string) (TokenClaims, error) {
	clm := TokenClaims{}
	_, err := jwt.ParseWithClaims(token, &clm, func(token *jwt.Token) (interface{}, error) {
		return []byte(a.secretKey), nil
	})

	return clm, err
}
