// Package server wires the Gin HTTP server: global middleware, health check,
// and the /api/v1 route group that feature modules attach handlers to.
package server

import (
	"time"

	_ "github.com/AbolfazlTayyar/gogym/docs/swagger"
	"github.com/AbolfazlTayyar/gogym/internal/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// APIV1Prefix is the route group under which every feature endpoint is registered.
const APIV1Prefix = "/api/v1"

// New builds a Gin engine with global middleware (recovery, structured request
// logging, CORS) and the /healthz and /api/v1 routes wired in.
func New(cfg config.Config, gormDB *gorm.DB, log zerolog.Logger) *gin.Engine {
	if cfg.Environment == config.EnvProduction {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery(), requestLogger(log))

	// TODO: lock down allowed origins once the frontend stack and its
	// deployed origin are chosen (frontend stack is still undecided per
	// docs/mvp-spec.md). AllowAllOrigins is permissive on purpose for now.
	router.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"Origin", "Content-Type", "Authorization"},
		MaxAge:          12 * time.Hour,
	}))

	router.GET("/healthz", healthzHandler(gormDB))

	if cfg.Environment == config.EnvDevelopment {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	router.Group(APIV1Prefix)

	return router
}
