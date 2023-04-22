package handler

import (
	"context"
	"errors"

	"github.com/bufbuild/connect-go"

	"github.com/ashep/a23n/api"
	"github.com/ashep/a23n/sdk/proto/a23n/v1"
)

func (h *Handler) UpdateEntity(
	ctx context.Context,
	req *connect.Request[v1.UpdateEntityRequest],
) (*connect.Response[v1.UpdateEntityResponse], error) {
	// TODO: authorize
	//crd, err := h.credentialsFromCtx(ctx)
	//if err != nil {
	//	return nil, err
	//}

	//if crd.Token != h.cfg.Secret {
	//	return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	//}

	err := h.api.UpdateEntity(ctx, req.Msg.Id, req.Msg.Secret, req.Msg.Scope, nil)
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
