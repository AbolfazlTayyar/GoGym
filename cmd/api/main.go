// Command api runs the gogym HTTP server.
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
