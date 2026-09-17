package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// healthzPingTimeout bounds how long the DB ping in healthzHandler may take,
// so an unreachable DB fails fast instead of hanging the request.
const healthzPingTimeout = 3 * time.Second

// healthzOKResponse is the body returned when the service and its database
// connection are healthy.
type healthzOKResponse struct {
	Status string `json:"status" example:"ok"`
}

// healthzErrorResponse is the body returned when the database is unreachable.
type healthzErrorResponse struct {
	Status string `json:"status" example:"unavailable"`
	Reason string `json:"reason" example:"failed to connect to database"`
}

// healthzHandler reports 200 with {"status":"ok"} when the DB is reachable,
// or 503 with the failure reason otherwise.
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
			c.JSON(http.StatusServiceUnavailable, healthzErrorResponse{Status: "unavailable", Reason: err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), healthzPingTimeout)
		defer cancel()

		if err := sqlDB.PingContext(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, healthzErrorResponse{Status: "unavailable", Reason: err.Error()})
			return
		}

		c.JSON(http.StatusOK, healthzOKResponse{Status: "ok"})
	}
}
