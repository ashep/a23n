package api

import (
	"context"
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func (a *API) Authenticate(ctx context.Context, id, secret string) (string, *jwt.NumericDate, error) {
	e, err := a.GetEntity(ctx, id)
	if err != nil {
		return "", nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(e.Secret), []byte(secret))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return "", nil, nil
	} else if err != nil {
		return "", nil, err
	}

	t := a.CreateToken(e)

	exp, err := t.Claims.GetExpirationTime()
	if err != nil {
		return "", nil, err
	}

	s, err := a.GetTokenSignedString(t)
	if err != nil {
		return "", nil, err
	}

	return s, exp, nil
}

func (a *API) Authorize(e Entity, scope []string) bool {
	eScope := make(map[string]bool)
	for _, s := range e.Scope {
		if s == "*" {
			return true
		}
		eScope[s] = true
	}

	for _, reqS := range scope {
		if !eScope[reqS] {
			return false
		}
	}

	return true
}
