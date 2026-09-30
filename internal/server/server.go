// Package server wires the Gin engine, global middleware, and routes.
package server

import (
	"time"

	_ "github.com/AbolfazlTayyar/gogym/docs/swagger"
	"github.com/AbolfazlTayyar/gogym/internal/coach"
	"github.com/AbolfazlTayyar/gogym/internal/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

const APIV1Prefix = "/api/v1"

func New(cfg config.Config, gormDB *gorm.DB, log zerolog.Logger) *gin.Engine {
	if cfg.Environment == config.EnvProduction {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	router.Use(gin.CustomRecovery(recoveryHandler), requestLogger(log))
	router.HandleMethodNotAllowed = true
	router.NoRoute(notFoundHandler)
	router.NoMethod(methodNotAllowedHandler)

	// TODO: restrict AllowAllOrigins once the frontend's deployed origin is known.
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

	coachRepo := coach.NewRepository(gormDB)
	coachSvc := coach.NewService(coachRepo, cfg.JWTSecret, cfg.JWTExpiry)
	coachHandler := coach.NewHandler(coachSvc, coachRepo)

	v1 := router.Group(APIV1Prefix)
	protected := v1.Group("")
	protected.Use(coach.AuthMiddleware(coachSvc))

	coach.RegisterRoutes(v1, protected, coachHandler)

	return router
}
