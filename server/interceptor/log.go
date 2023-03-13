package interceptor

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rs/zerolog"
)

func Log(l zerolog.Logger) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			l.Debug().
				Str("addr", req.Peer().Addr).
				Str("proto", req.Peer().Protocol).
				Str("proc", req.Spec().Procedure).
				Msg("request")
			return next(ctx, req)
		}
	}
}
