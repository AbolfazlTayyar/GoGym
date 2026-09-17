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

// healthzHandler reports 200 with {"status":"ok"} when the DB is reachable,
// or 503 with the failure reason otherwise.
func healthzHandler(gormDB *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		sqlDB, err := gormDB.DB()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "reason": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), healthzPingTimeout)
		defer cancel()

		if err := sqlDB.PingContext(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "reason": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
