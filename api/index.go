// Package handler runs the API as a Vercel Go function; leaving Vercel removes it (docs/adr/0039).
package handler

import (
	"net/http"
	"os"

	"github.com/AbolfazlTayyar/gogym/internal/platform/config"
	"github.com/AbolfazlTayyar/gogym/internal/platform/db"
	"github.com/AbolfazlTayyar/gogym/internal/server"
	"github.com/rs/zerolog"
)

// vercelClientIPHeader is set by Vercel's edge to the caller's IP, overwriting any value the caller sent.
const vercelClientIPHeader = "X-Real-IP"

var router http.Handler

// init builds the router once per function instance, so warm requests reuse it and its DB pool.
func init() {
	log := zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg := config.Load()
	log.Info().Str("config", cfg.String()).Msg("config loaded successfully")

	gormDB, err := db.New(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}

	engine, err := server.New(cfg, gormDB, log)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to build router")
	}
	// The peer is Vercel's runtime, not the caller, so TRUSTED_PROXIES can't name it; only this
	// entry point may trust the header, since anywhere else a caller could set it.
	engine.TrustedPlatform = vercelClientIPHeader
	router = engine
}

func Handler(w http.ResponseWriter, r *http.Request) {
	router.ServeHTTP(w, r)
}
