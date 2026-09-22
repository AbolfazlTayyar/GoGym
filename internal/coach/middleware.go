package coach

import (
	"net/http"
	"strings"

	"github.com/AbolfazlTayyar/gogym/internal/httpx"
	"github.com/AbolfazlTayyar/gogym/internal/tenant"
	"github.com/gin-gonic/gin"
)

const bearerPrefix = "Bearer "

// AuthMiddleware validates the bearer JWT on protected routes and, on
// success, records the authenticated coach's id via tenant.SetCoachID for
// handlers and repositories to read. It rejects with 401 if the header is
// missing, malformed, or the token is invalid or expired.
func AuthMiddleware(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, bearerPrefix) {
			httpx.Error(c, http.StatusUnauthorized, httpx.CodeUnauthorized, httpx.MsgUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(header, bearerPrefix)

		coachID, err := svc.ValidateToken(tokenString)
		if err != nil {
			httpx.Error(c, http.StatusUnauthorized, httpx.CodeUnauthorized, httpx.MsgUnauthorized)
			return
		}

		tenant.SetCoachID(c, coachID)
		c.Next()
	}
}
