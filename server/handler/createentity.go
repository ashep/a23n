package handler

import (
	"context"
	"errors"

	"github.com/bufbuild/connect-go"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/ashep/a23n/api"
	v1 "github.com/ashep/a23n/sdk/proto/a23n/v1"
)

func (h *Handler) CreateEntity(
	ctx context.Context,
	req *connect.Request[v1.CreateEntityRequest],
) (*connect.Response[v1.CreateEntityResponse], error) {
	// TODO: authorize
	//crd, err := h.credentialsFromCtx(ctx)
	//if err != nil {
	//	return nil, err
	//}

	//if crd.Token != h.Secret {
	//	return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	//}

	id := uuid.NewString()

	secretHash, err := bcrypt.GenerateFromPassword([]byte(req.Msg.Secret), bcrypt.DefaultCost)
	if err != nil {
		h.l.Error().Err(err).Msg("generate password hash")
		return nil, connect.NewError(connect.CodeInternal, nil)
	}

	err = h.api.CreateEntity(ctx, id, secretHash, req.Msg.Scope, req.Msg.Attrs)
	if errors.Is(err, api.ErrInvalidArg{}) {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	} else if err != nil {
		h.l.Error().Err(err).Msg("create entity")
		return nil, connect.NewError(connect.CodeInternal, nil)
	}

	h.l.Info().
		Str("id", id).
		Strs("scope", req.Msg.Scope).
		Interface("attrs", req.Msg.Attrs).
		Msg("entity created")

	return connect.NewResponse(&v1.CreateEntityResponse{Id: id}), nil
}
