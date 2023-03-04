package handler

import (
	"context"

	"github.com/bufbuild/connect-go"

	"github.com/ashep/a23n/sdk/proto/a23n/v1"
)

func (h *Handler) Authenticate(ctx context.Context, _ *connect.Request[v1.AuthenticateRequest],
) (*connect.Response[v1.AuthenticateResponse], error) {
	crd, err := h.getCredentialsFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	if crd.ID != "" && crd.Password != "" {
		tok, exp, err := h.api.Authenticate(ctx, crd.ID, crd.Password)
		if err != nil {
			h.l.Error().Err(err).Str("entity_id", crd.ID).Msg("authenticate failed")
			return nil, connect.NewError(connect.CodeUnauthenticated, nil)
		}

		h.l.Debug().Str("entity_id", crd.ID).Msg("authenticated")

		return connect.NewResponse(&v1.AuthenticateResponse{Token: tok, Expires: exp}), nil
	}

	if crd.Token != "" {
		e, err := h.api.GetEntityByTokenString(ctx, crd.Token)
		if err != nil {
			h.l.Error().Err(err).Msg("get entity by token")
			return nil, connect.NewError(connect.CodeUnauthenticated, nil)
		}

		t := h.api.CreateToken(e)

		ts, err := h.api.GetTokenSignedString(t)
		if err != nil {
			h.l.Error().Err(err).Msg("get token signed string")
			return nil, connect.NewError(connect.CodeUnauthenticated, nil)
		}

		exp, err := t.Claims.GetExpirationTime()
		if err != nil {
			h.l.Error().Err(err).Msg("get token expiration time")
			return nil, connect.NewError(connect.CodeUnauthenticated, nil)
		}

		h.l.Debug().Str("entity_id", e.ID).Msg("authenticated")

		return connect.NewResponse(&v1.AuthenticateResponse{Token: ts, Expires: exp.Unix()}), nil
	}

	return nil, connect.NewError(connect.CodeUnauthenticated, nil)
}

func (h *Handler) Authorize(ctx context.Context, req *connect.Request[v1.AuthorizeRequest],
) (*connect.Response[v1.AuthorizeResponse], error) {
	crd, err := h.getCredentialsFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	e, err := h.api.GetEntityByTokenString(ctx, crd.Token)
	if err != nil {
		h.l.Error().Err(err).Msg("get entity by token")
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	if !h.api.Authorize(e, req.Msg.Attrs) {
		h.l.Error().Err(err).Msg("not authorized")
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	h.l.Debug().Str("entity_id", e.ID).Msg("authorized")

	return connect.NewResponse(&v1.AuthorizeResponse{}), nil
}
