package handler

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rs/zerolog"

	"github.com/ashep/a23n/api"
	"github.com/ashep/a23n/server/credentials"
)

type Handler struct {
	api *api.API
	l   zerolog.Logger
}

func New(svc *api.API, l zerolog.Logger) *Handler {
	return &Handler{
		api: svc,
		l:   l,
	}
}

func (h *Handler) getCredentialsFromCtx(ctx context.Context) (credentials.Credentials, error) {
	crd, ok := ctx.Value("crd").(credentials.Credentials)
	if !ok {
		return crd, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	return crd, nil
}
