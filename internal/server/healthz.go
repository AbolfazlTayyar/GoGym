package server

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	healthzPath        = "/healthz"
	healthzPingTimeout = 3 * time.Second
	// Caps the endpoint at one database ping per window, however often it's hit.
	healthzCacheTTL = 5 * time.Second

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

// healthChecker remembers the last ping's result for ttl.
type healthChecker struct {
	ping func(context.Context) error
	ttl  time.Duration
	now  func() time.Time

	mu        sync.Mutex
	checkedAt time.Time
	lastErr   error
}

// check holds the lock across the ping so concurrent callers share one ping instead of each starting their own.
func (h *healthChecker) check(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.now().Sub(h.checkedAt) < h.ttl {
		return h.lastErr
	}

	// Detached from the caller so a client hanging up mid-ping isn't cached as an outage.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), healthzPingTimeout)
	defer cancel()

	h.lastErr = h.ping(ctx)
	h.checkedAt = h.now()
	return h.lastErr
}

// healthzHandler is deliberately unenveloped: infra probes match its exact bare body.
//
// @Summary		Health check
// @Description	Reports whether the service and its database connection are healthy. The result is cached for a few seconds.
// @Tags			health
// @Produce		json
// @Success		200	{object}	healthzOKResponse
// @Failure		503	{object}	healthzErrorResponse
// @Router			/healthz [get]
func healthzHandler(gormDB *gorm.DB) gin.HandlerFunc {
	return newHealthzHandler(&healthChecker{
		ping: func(ctx context.Context) error {
			sqlDB, err := gormDB.DB()
			if err != nil {
				return err
			}
			return sqlDB.PingContext(ctx)
		},
		ttl: healthzCacheTTL,
		now: time.Now,
	})
}

func newHealthzHandler(checker *healthChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := checker.check(c.Request.Context()); err != nil {
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
