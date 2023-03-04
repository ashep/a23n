package api

import (
	"database/sql"

	"github.com/rs/zerolog"
)

type API struct {
	db       *sql.DB
	secret   string
	tokenTTL int
	l        zerolog.Logger

	tokenCache  map[string]string
	entityCache map[string]Entity
}

func New(db *sql.DB, secret string, tokenTTL int, l zerolog.Logger) *API {
	if secret == "" {
		secret = randString(32)
		l.Warn().Str("secret", secret).Msg("random secret generated, consider putting it to the config file")
	}

	if tokenTTL == 0 {
		tokenTTL = 86400
	}
	l.Debug().Int("token_ttl", tokenTTL).Msg("")

	return &API{
		db:       db,
		secret:   secret,
		tokenTTL: tokenTTL,
		l:        l,

		tokenCache:  make(map[string]string),
		entityCache: make(map[string]Entity),
	}
}
