package api

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

func (a *API) Authenticate(ctx context.Context, id, secret string) (string, int64, error) {
	e, err := a.GetEntity(ctx, id)
	if err != nil {
		return "", 0, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(e.Secret), []byte(secret))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return "", 0, ErrUnauthorized
	} else if err != nil {
		return "", 0, err
	}

	t := a.CreateToken(e)

	exp, err := t.Claims.GetExpirationTime()
	if err != nil {
		return "", 0, err
	}

	s, err := a.GetTokenSignedString(t)
	if err != nil {
		return "", 0, err
	}

	return s, exp.Unix(), nil
}

func (a *API) Authorize(e Entity, attrs []string) bool {
	eAttrs := make(map[string]bool)
	for _, attr := range e.Attrs {
		if attr == "*" {
			return true
		}
		eAttrs[attr] = true
	}

	for _, reqAttr := range attrs {
		if !eAttrs[reqAttr] {
			return false
		}
	}

	return true
}
