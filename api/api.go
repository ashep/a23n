package api

import (
	"database/sql"
	"errors"
)

type API struct {
	db       *sql.DB
	secret   string
	tokenTTL int
}

func New(db *sql.DB, secret string, tokenTTL int) (*API, error) {
	if len(secret) < 32 {
		return nil, errors.New("secret key is too short")
	}

	if tokenTTL == 0 {
		tokenTTL = 86400
	}

	return &API{
		db:       db,
		secret:   secret,
		tokenTTL: tokenTTL,
	}, nil
}
