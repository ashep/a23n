package handler

import (
	"context"
	"time"

	"github.com/rs/zerolog"

	"github.com/ashep/a23n/api"
	"github.com/ashep/a23n/server/credentials"
)

type Handler struct {
	api             api.API
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	l               zerolog.Logger
}

func New(api api.API, accessTokenTTL, refreshTokenTTL time.Duration, l zerolog.Logger) *Handler {
	return &Handler{
		api:             api,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
		l:               l,
	}
}

func (h *Handler) credentialsFromCtx(ctx context.Context) (credentials.Credentials, bool) {
	crd, ok := ctx.Value("crd").(credentials.Credentials)
	return crd, ok
}
