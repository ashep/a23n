package server

import (
	"context"
	"net/http"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/rs/zerolog"

	"github.com/ashep/a23n/api"
	"github.com/ashep/a23n/sdk/proto/a23n/v1/v1connect"
	"github.com/ashep/a23n/server/handler"
	"github.com/ashep/a23n/server/interceptor"
)

type Server struct {
	api             api.API
	addr            string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	l               zerolog.Logger
}

func New(api api.API, addr string, accessTokenTTL, refreshTokenTTL time.Duration, l zerolog.Logger) *Server {
	return &Server{
		api:             api,
		addr:            addr,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
		l:               l,
	}
}

func corsHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, ResponseType")

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}

func (s *Server) Run(ctx context.Context) error {
	interceptors := connect.WithInterceptors(
		interceptor.Auth(s.l),
		interceptor.Log(s.l),
	)

	p, h := v1connect.NewAuthServiceHandler(handler.New(s.api, s.accessTokenTTL, s.refreshTokenTTL, s.l), interceptors)

	mux := http.NewServeMux()
	mux.Handle(p, corsHandler(h))

	srv := &http.Server{Addr: s.addr, Handler: mux}

	go func() {
		<-ctx.Done()
		if errF := srv.Close(); errF != nil {
			s.l.Error().Err(errF).Msg("failed to close server")
		}
	}()

	s.l.Debug().Str("addr", s.addr).Msg("starting server")
	return srv.ListenAndServe()
}
