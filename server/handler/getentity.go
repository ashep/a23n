package handler

import (
	"context"

	"github.com/bufbuild/connect-go"

	v1 "github.com/ashep/a23n/sdk/proto/a23n/v1"
)

func (h *Handler) GetEntity(
	ctx context.Context,
	_ *connect.Request[v1.GetEntityRequest],
) (*connect.Response[v1.GetEntityResponse], error) {
	crd, ok := h.credentialsFromCtx(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	e, err := h.api.GetEntity(ctx, crd.Token)
	if err != nil {
		h.l.Warn().Err(err).Msg("get entity by token")
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	return connect.NewResponse(&v1.GetEntityResponse{Id: e.ID, Scope: e.Scope}), nil
}
