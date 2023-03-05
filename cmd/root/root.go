package root

import (
	"database/sql"
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

			db, err := sql.Open("postgres", cfg.DB.DSN)
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

			apiSec := os.Getenv("A23N_API_SECRET")
			if apiSec != "" {
				cfg.API.Secret = apiSec
			}

			apiTokenTTL := os.Getenv("A23N_API_TOKEN_TTL")
			if apiTokenTTL != "" {
				t, _ := strconv.Atoi(apiTokenTTL)
				cfg.API.TokenTTL = t
			}

			a := api.New(db, cfg.API.Secret, cfg.API.TokenTTL, l.With().Str("pkg", "api").Logger())
			s := server.New(cfg.Server, a, l.With().Str("pkg", "server").Logger())
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
