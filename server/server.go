package server

import (
	"context"
	"net/http"

	"github.com/ashep/a23n/config"
	"github.com/bufbuild/connect-go"
	"github.com/rs/zerolog"

	"github.com/ashep/a23n/api"
	"github.com/ashep/a23n/sdk/proto/a23n/v1/v1connect"
	"github.com/ashep/a23n/server/handler"
	"github.com/ashep/a23n/server/interceptor"
)

type Server struct {
	cfg config.Server
	svc *api.API
	l   zerolog.Logger
}

func New(cfg config.Server, svc *api.API, l zerolog.Logger) *Server {
	if cfg.Address == "" {
		cfg.Address = "localhost:9000"
	}

	return &Server{
		cfg: cfg,
		svc: svc,
		l:   l,
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

	p, h := v1connect.NewAuthServiceHandler(handler.New(s.cfg, s.svc, s.l), interceptors)

	mux := http.NewServeMux()
	mux.Handle(p, corsHandler(h))

	srv := &http.Server{Addr: s.cfg.Address, Handler: mux}

	go func() {
		<-ctx.Done()
		if errF := srv.Close(); errF != nil {
			s.l.Error().Err(errF).Msg("failed to close server")
		}
	}()

	s.l.Debug().Str("addr", s.cfg.Address).Msg("starting server")
	return srv.ListenAndServe()
}
