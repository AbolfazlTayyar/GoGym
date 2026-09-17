package server

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// requestLogger emits one structured JSON log line per request to stdout,
// per docs/prestart-roadmap.md step 10.
func requestLogger(log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		event := log.Info()
		if len(c.Errors) > 0 {
			event = log.Error()
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
