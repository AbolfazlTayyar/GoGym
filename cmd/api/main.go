// Command api runs the gogym HTTP server.
//
// @title			GoGym API
// @version			1.0
// @description		API for the GoGym coaching platform.
//
// BasePath is root, not /api/v1: /healthz is intentionally mounted outside
// the versioned API group (unversioned infra probe), so a global /api/v1
// basePath would make Swagger UI's "Try it out" call the wrong URL for it.
// Every handler's own @Router annotation must therefore spell out its full
// path, e.g. "@Router /api/v1/coaches [post]" for versioned endpoints and
// "@Router /healthz [get]" for unversioned ones.
// @BasePath		/
//
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description					Type "Bearer" followed by a space and a JWT token.
package main

import (
	"os"

	"github.com/AbolfazlTayyar/gogym/internal/config"
	"github.com/AbolfazlTayyar/gogym/internal/db"
	"github.com/AbolfazlTayyar/gogym/internal/server"
	"github.com/rs/zerolog"
)

func main() {
	log := zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg := config.Load()
	log.Info().Str("config", cfg.String()).Msg("config loaded successfully")

	gormDB, err := db.New(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}

	router := server.New(cfg, gormDB, log)

	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal().Err(err).Msg("server stopped")
	}
}
