// Package server wires the Gin engine, global middleware, and routes.
package server

import (
	"fmt"

	_ "github.com/AbolfazlTayyar/gogym/docs/swagger"
	"github.com/AbolfazlTayyar/gogym/internal/athlete"
	"github.com/AbolfazlTayyar/gogym/internal/coach"
	"github.com/AbolfazlTayyar/gogym/internal/measurement"
	"github.com/AbolfazlTayyar/gogym/internal/movement"
	"github.com/AbolfazlTayyar/gogym/internal/plan"
	"github.com/AbolfazlTayyar/gogym/internal/platform/config"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

const APIV1Prefix = "/api/v1"

func New(cfg config.Config, gormDB *gorm.DB, log zerolog.Logger) (*gin.Engine, error) {
	if cfg.Environment == config.EnvProduction {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Gin trusts X-Forwarded-For from any peer by default, which lets a caller pick its own
	// ClientIP and dodge the per-IP auth rate limit.
	if err := router.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		return nil, fmt.Errorf("server: invalid trusted proxies: %w", err)
	}

	// The logger wraps recovery so a recovered panic still gets its request log line.
	router.Use(requestLogger(log), gin.CustomRecovery(recoveryHandler))
	router.HandleMethodNotAllowed = true
	router.NoRoute(notFoundHandler)
	router.NoMethod(methodNotAllowedHandler)

	// No CORS middleware: the frontend is served on the API's origin (docs/adr/0030), so no other
	// origin may call the API with the coach's cookie — see docs/adr/0031.

	router.GET(healthzPath, healthzHandler(gormDB))

	if cfg.Environment == config.EnvDevelopment {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	coachSvc := coach.NewService(coach.NewRepository(gormDB), cfg.JWTSecret, cfg.JWTExpiry)
	coachHandler := coach.NewHandler(coachSvc, cfg.Environment != config.EnvDevelopment)

	v1 := router.Group(APIV1Prefix)
	protected := v1.Group("")
	protected.Use(coach.AuthMiddleware(coachSvc))

	coach.RegisterRoutes(v1, protected, coachHandler)

	athleteRepo := athlete.NewRepository(gormDB)
	athleteSvc := athlete.NewService(athleteRepo)
	athlete.RegisterRoutes(protected, athlete.NewHandler(athleteSvc))

	measurementRepo := measurement.NewRepository(gormDB)
	measurementSvc := measurement.NewService(measurementRepo, athleteSvc)
	measurement.RegisterRoutes(protected, measurement.NewHandler(measurementSvc))

	movementRepo := movement.NewRepository(gormDB)
	movementSvc := movement.NewService(movementRepo)
	movement.RegisterRoutes(protected, movement.NewHandler(movementSvc))

	planRepo := plan.NewRepository(gormDB)
	planSvc := plan.NewService(planRepo, athleteSvc, movementSvc)
	plan.RegisterRoutes(protected, plan.NewHandler(planSvc))

	return router, nil
}
