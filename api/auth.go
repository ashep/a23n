package api

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

func (a *DefaultAPI) CheckSecret(hashed, secret string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(secret))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}
