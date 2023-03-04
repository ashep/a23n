package interceptor

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/ashep/a23n/server/credentials"
	"github.com/bufbuild/connect-go"
	"github.com/rs/zerolog"
)

func Auth(l zerolog.Logger) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			crd := credentials.Credentials{}
			authHdr := req.Header().Get("Authorization")

			if strings.HasPrefix(authHdr, "Basic") {
				basicStr := strings.TrimPrefix(authHdr, "Basic ")
				authB, err := base64.StdEncoding.DecodeString(basicStr)
				if err != nil {
					l.Warn().Err(err).Msg("failed to decode authorization header")
					return nil, connect.NewError(connect.CodeUnauthenticated, nil)
				}
				basicStr = string(authB)

				basicSplit := strings.Split(basicStr, ":")
				if len(basicSplit) != 2 {
					return nil, connect.NewError(connect.CodeUnauthenticated, nil)
				}

				crd.ID = basicSplit[0]
				crd.Password = basicSplit[1]
			} else if strings.HasPrefix(authHdr, "Bearer") {
				crd.Token = strings.TrimPrefix(authHdr, "Bearer ")
			}

			return next(context.WithValue(ctx, "crd", crd), req)
		}
	}
}
