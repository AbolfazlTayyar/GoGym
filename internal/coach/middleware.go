package coach

import (
	"net/http"
	"strings"

	"github.com/AbolfazlTayyar/gogym/internal/platform/httpx"
	"github.com/AbolfazlTayyar/gogym/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

const bearerPrefix = "Bearer "

// AuthMiddleware takes the token from the browser's cookie, or from the Bearer header that Swagger UI's
// Authorize, curl and other non-browser clients send.
func AuthMiddleware(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := requestToken(c)
		if tokenString == "" {
			httpx.Error(c, http.StatusUnauthorized, httpx.CodeUnauthorized, httpx.MsgUnauthorized)
			return
		}

		coachID, err := svc.ValidateToken(tokenString)
		if err != nil {
			httpx.Error(c, http.StatusUnauthorized, httpx.CodeUnauthorized, httpx.MsgUnauthorized)
			return
		}

		tenant.SetCoachID(c, coachID)
		c.Next()
	}
}

func requestToken(c *gin.Context) string {
	if cookie, err := c.Cookie(AuthCookieName); err == nil && cookie != "" {
		return cookie
	}
	if header := c.GetHeader("Authorization"); strings.HasPrefix(header, bearerPrefix) {
		return strings.TrimPrefix(header, bearerPrefix)
	}
	return ""
}
