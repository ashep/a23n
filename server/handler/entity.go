package handler

import (
	"context"
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

	return connect.NewResponse(&v1.GetEntityResponse{Id: e.ID, Scope: e.Scope}), nil
}

func (h *Handler) CreateEntity(
	ctx context.Context,
	req *connect.Request[v1.CreateEntityRequest],
) (*connect.Response[v1.CreateEntityResponse], error) {
	crd, err := h.getCredentialsFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	if crd.Token != h.cfg.Secret {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	id, err := h.api.CreateEntity(ctx, req.Msg.Secret, req.Msg.Note, req.Msg.Scope)
	if errors.Is(err, api.ErrInvalidArg{}) {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	} else if err != nil {
		h.l.Error().Err(err).Msg("create entity")
		return nil, connect.NewError(connect.CodeInternal, nil)
	}

	h.l.Info().
		Str("id", id).
		Strs("scope", req.Msg.Scope).
		Str("note", req.Msg.Note).
		Msg("entity created")

	return connect.NewResponse(&v1.CreateEntityResponse{Id: id}), nil
}

func (h *Handler) UpdateEntity(
	ctx context.Context,
	req *connect.Request[v1.UpdateEntityRequest],
) (*connect.Response[v1.UpdateEntityResponse], error) {
	crd, err := h.getCredentialsFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	if crd.Token != h.cfg.Secret {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	err = h.api.UpdateEntity(ctx, req.Msg.Id, req.Msg.Secret, req.Msg.Note, req.Msg.Scope)
	if errors.Is(err, api.ErrInvalidArg{}) {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	} else if errors.Is(err, api.ErrNotFound) {
		return nil, connect.NewError(connect.CodeNotFound, err)
	} else if err != nil {
		h.l.Error().Err(err).Msg("update entity")
		return nil, connect.NewError(connect.CodeInternal, nil)
	}

	h.l.Info().
		Str("id", req.Msg.Id).
		Strs("scope", req.Msg.Scope).
		Str("note", req.Msg.Note).
		Msg("entity updated")

	return connect.NewResponse(&v1.UpdateEntityResponse{Id: req.Msg.Id}), nil
}
