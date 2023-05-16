package handler

import (
	"context"

	"github.com/bufbuild/connect-go"

	v1 "github.com/ashep/a23n/sdk/proto/a23n/v1"
)

func (h *Handler) RefreshToken(
	ctx context.Context,
	req *connect.Request[v1.RefreshTokenRequest],
) (*connect.Response[v1.RefreshTokenResponse], error) {
	crd, ok := h.credentialsFromCtx(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	if crd.Token == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	e, err := h.api.GetEntity(ctx, "") // FIXME: wip
	if err != nil {
		h.l.Error().Err(err).Msg("api.GetEntityByToken failed")
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	// TODO: pass scope
	t := h.api.CreateToken(e.ID, []string{}, 123)
	ts, err := h.api.GetTokenSignedString(t)
	if err != nil {
		h.l.Error().Err(err).Str("entity_id", e.ID).Msg("api.GetTokenSignedString failed")
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	exp, err := t.Claims().GetExpirationTime()
	if err != nil {
		h.l.Error().Err(err).Str("entity_id", e.ID).Msg("token GetExpirationTime failed")
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	h.l.Info().
		Str("entity_id", e.ID).
		Str("expires", exp.String()).
		Msg("authenticated by token")

	return connect.NewResponse(&v1.RefreshTokenResponse{Token: ts, TokenExpires: exp.Unix()}), nil

}
