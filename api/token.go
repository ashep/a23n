package api

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenClaims struct {
	jwt.RegisteredClaims
	Scope []string `json:"scope,omitempty"`
}

func (a *DefaultAPI) CreateToken(subject string, scope []string, ttl time.Duration) *jwt.Token {
	n := jwt.NewNumericDate(time.Now())

	return jwt.NewWithClaims(jwt.SigningMethodHS256, TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			IssuedAt:  n,
			NotBefore: n,
			ExpiresAt: jwt.NewNumericDate(n.Add(ttl)),
		},
		Scope: scope,
	})
}

func (a *DefaultAPI) GetTokenSignedString(t *jwt.Token) (string, error) {
	return t.SignedString([]byte(a.secret))
}

func (a *DefaultAPI) ParseToken(token string) (TokenClaims, error) {
	clm := TokenClaims{}
	_, err := jwt.ParseWithClaims(token, &clm, func(token *jwt.Token) (interface{}, error) {
		return []byte(a.secret), nil
	})

	return clm, err
}
