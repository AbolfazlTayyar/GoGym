package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	healthzPath        = "/healthz"
	healthzPingTimeout = 3 * time.Second

	healthzStatusUnavailable = "unavailable"
	// The real error goes to the logs; the endpoint is public, and driver errors name internal hosts.
	healthzReasonDBUnreachable = "database unreachable"
)

type healthzOKResponse struct {
	Status string `json:"status" example:"ok"`
}

type healthzErrorResponse struct {
	Status string `json:"status" example:"unavailable"`
	Reason string `json:"reason" example:"database unreachable"`
}

// healthzHandler is deliberately unenveloped: infra probes match its exact bare body.
//
// @Summary		Health check
// @Description	Reports whether the service and its database connection are healthy.
// @Tags			health
// @Produce		json
// @Success		200	{object}	healthzOKResponse
// @Failure		503	{object}	healthzErrorResponse
// @Router			/healthz [get]
func healthzHandler(gormDB *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		sqlDB, err := gormDB.DB()
		if err != nil {
			healthzUnavailable(c, err)
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), healthzPingTimeout)
		defer cancel()

		if err := sqlDB.PingContext(ctx); err != nil {
			healthzUnavailable(c, err)
			return
		}

		c.JSON(http.StatusOK, healthzOKResponse{Status: "ok"})
	}
}

func healthzUnavailable(c *gin.Context, err error) {
	_ = c.Error(err) // requestLogger writes it out at error level
	c.JSON(http.StatusServiceUnavailable, healthzErrorResponse{Status: healthzStatusUnavailable, Reason: healthzReasonDBUnreachable})
}
