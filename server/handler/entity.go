package handler

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/bufbuild/connect-go"

	"github.com/ashep/a23n/api"
	"github.com/ashep/a23n/sdk/proto/a23n/v1"
)

func (h *Handler) GetEntity(
	ctx context.Context,
	_ *connect.Request[v1.GetEntityRequest],
) (*connect.Response[v1.GetEntityResponse], error) {
	crd, err := h.getCredentialsFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	e, err := h.api.GetEntityByTokenString(ctx, crd.Token)
	if err != nil {
		h.l.Warn().Err(err).Msg("get entity by token")
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	return connect.NewResponse(&v1.GetEntityResponse{Id: e.ID, Attrs: e.Attrs}), nil
}

func (h *Handler) CreateEntity(
	ctx context.Context,
	req *connect.Request[v1.CreateEntityRequest],
) (*connect.Response[v1.CreateEntityResponse], error) {
	crd, err := h.getCredentialsFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	entitySec := req.Msg.Secret
	if len(entitySec) < 8 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("secret is too short"))
	}

	if req.Msg.Attrs == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("empty attrs"))
	}

	attrs := make([]string, 0)
	if err := json.Unmarshal([]byte(req.Msg.Attrs), &attrs); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to unmarshal attrs"))
	}

	e, err := h.api.CreateEntity(ctx, crd.Token, entitySec, attrs, req.Msg.Note)
	if errors.Is(err, api.ErrUnauthorized) {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	} else if errors.Is(err, api.ErrInvalidArg{}) {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	} else if err != nil {
		h.l.Error().Err(err).Msg("create entity")
		return nil, connect.NewError(connect.CodeInternal, nil)
	}

	return connect.NewResponse(&v1.CreateEntityResponse{Id: e.ID}), nil
}
