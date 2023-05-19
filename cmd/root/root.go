package root

import (
	"errors"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/ashep/a23n/api"
	"github.com/ashep/a23n/config"
	"github.com/ashep/a23n/logger"
	"github.com/ashep/a23n/migration"
	"github.com/ashep/a23n/server"
	"github.com/ashep/a23n/sqldb"
)

var (
	debugMode  bool
	configPath string
	migUp      bool
	migDown    bool
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			var err error

			rand.Seed(time.Now().UnixNano())

			if !debugMode && os.Getenv("A23N_DEBUG") == "1" {
				debugMode = true
			}
			l := logger.New(debugMode)

			cfg := config.Config{}
			if configPath != "" {
				cfg, err = config.ParseFromPath(configPath)
				if err != nil {
					l.Fatal().Err(err).Msg("failed to load config")
					return
				}
			}

			dbDSN := os.Getenv("A23N_DB_DSN")
			if dbDSN != "" {
				cfg.DB.DSN = dbDSN
			}
			if cfg.DB.DSN == "" {
				l.Fatal().Err(err).Msg("empty db dsn")
				return
			}

			db, err := sqldb.NewPostgres(cfg.DB.DSN)
			if err != nil {
				l.Fatal().Err(err).Msg("failed to open db")
				return
			}

			if err = db.PingContext(cmd.Context()); err != nil {
				l.Fatal().Err(err).Msg("failed to connect to db")
			}
			l.Debug().Msg("db connection ok")

			if migUp {
				if err := migration.Up(db); err != nil {
					l.Fatal().Err(err).Msg("failed to apply migrations")
				}
				return
			}

			if migDown {
				if err := migration.Down(db); err != nil {
					l.Fatal().Err(err).Msg("failed to revert migrations")
				}
				return
			}

			secret := os.Getenv("A23N_SECRET")
			if secret != "" {
				cfg.Secret = secret
			}

			accessTokenTTL := os.Getenv("A23N_ACCESS_TOKEN_TTL")
			if accessTokenTTL != "" {
				t, _ := strconv.Atoi(accessTokenTTL)
				cfg.AccessTokenTTL = uint(t)
			}

			refreshTokenTTL := os.Getenv("A23N_REFRESH_TOKEN_TTL")
			if refreshTokenTTL != "" {
				t, _ := strconv.Atoi(refreshTokenTTL)
				cfg.RefreshTokenTTL = uint(t)
			}

			a := api.NewDefault(db, cfg.Secret, time.Now)

			addr := os.Getenv("A23N_ADDRESS")
			if addr != "" {
				cfg.Address = addr
			}
			s := server.New(
				a,
				cfg.Address,
				time.Duration(cfg.AccessTokenTTL)*time.Second,
				time.Duration(cfg.RefreshTokenTTL)*time.Second,
				l.With().Str("pkg", "server").Logger(),
			)

			if err := s.Run(cmd.Context()); errors.Is(err, http.ErrServerClosed) {
				l.Info().Msg("server stopped")
			} else if err != nil {
				l.Error().Err(err).Msg("")
			}
		},
	}

	cmd.Flags().BoolVar(&migUp, "migrate-up", false, "apply database migrations")
	cmd.Flags().BoolVar(&migDown, "migrate-down", false, "revert database migrations")

	cmd.PersistentFlags().BoolVarP(&debugMode, "debug", "d", false, "enable debug mode")
	cmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "path to the config file")

	return cmd
}
