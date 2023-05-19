package handler

import (
	"context"
	"errors"

	"github.com/bufbuild/connect-go"

	"github.com/ashep/a23n/api"
	"github.com/ashep/a23n/sdk/proto/a23n/v1"
)

func (h *Handler) Authenticate(
	ctx context.Context,
	req *connect.Request[v1.AuthenticateRequest],
) (*connect.Response[v1.AuthenticateResponse], error) {
	crd, ok := h.credentialsFromCtx(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	if crd.ID == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("empty entity id"))
	}

	e, err := h.api.GetEntity(ctx, crd.ID)
	if errors.Is(err, api.ErrNotFound) {
		h.l.Warn().Str("entity_id", crd.ID).Msg("entity not found")
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	} else if err != nil {
		h.l.Error().Err(err).Str("entity_id", crd.ID).Msg("failed to get entity")
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	ok, err = h.api.CheckSecret(crd.ID, crd.Password)
	if err != nil {
		h.l.Error().Err(err).Str("entity_id", crd.ID).Msg("check secret failed")
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	} else if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	if !h.api.CheckScope(e.Scope, req.Msg.Scope) {
		return nil, connect.NewError(connect.CodePermissionDenied, nil)
	}

	accessToken := h.api.CreateToken(e.ID, e.Scope, h.accessTokenTTL)
	accessTokenExp, err := accessToken.Claims().GetExpirationTime()
	if err != nil {
		h.l.Error().Err(err).Str("entity_id", crd.ID).Msg("get access token expiration time failed")
		return nil, connect.NewError(connect.CodeInternal, nil)
	}
	accessTokenStr, err := accessToken.SignedString(h.api.SecretKey())
	if err != nil {
		h.l.Error().Err(err).Str("entity_id", crd.ID).Msg("get access token signed string failed")
		return nil, connect.NewError(connect.CodeInternal, nil)
	}

	refreshToken := h.api.CreateToken(e.ID+"_refresh", e.Scope, h.refreshTokenTTL)
	refreshTokenExp, err := refreshToken.Claims().GetExpirationTime()
	if err != nil {
		h.l.Error().Err(err).Str("entity_id", crd.ID).Msg("get refresh token expiration time failed")
		return nil, connect.NewError(connect.CodeInternal, nil)
	}
	refreshTokenStr, err := refreshToken.SignedString(h.api.SecretKey())
	if err != nil {
		h.l.Error().Err(err).Str("entity_id", crd.ID).Msg("get refresh token signed string failed")
		return nil, connect.NewError(connect.CodeInternal, nil)
	}

	h.l.Info().
		Str("entity_id", crd.ID).
		Int64("access_token_expires", accessTokenExp.Unix()).
		Int64("refresh_token_expires", refreshTokenExp.Unix()).
		Msg("authenticated by password")

	return connect.NewResponse(&v1.AuthenticateResponse{
		AccessToken:         accessTokenStr,
		AccessTokenExpires:  accessTokenExp.Unix(),
		RefreshToken:        refreshTokenStr,
		RefreshTokenExpires: refreshTokenExp.Unix(),
	}), nil
}
