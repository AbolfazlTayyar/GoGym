package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func requestLogger(log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		// Probes hit this every few seconds; only a failing check is worth a log line.
		if path == healthzPath && c.Writer.Status() == http.StatusOK {
			return
		}

		event := log.Info()
		if len(c.Errors) > 0 {
			event = log.Error().Strs("errors", c.Errors.Errors())
		}

		event.
			Str("method", c.Request.Method).
			Str("path", path).
			Str("query", query).
			Int("status", c.Writer.Status()).
			Dur("latency", time.Since(start)).
			Str("client_ip", c.ClientIP()).
			Msg("request handled")
	}
}
