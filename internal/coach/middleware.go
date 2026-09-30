package coach

import (
	"net/http"
	"strings"

	"github.com/AbolfazlTayyar/gogym/internal/httpx"
	"github.com/AbolfazlTayyar/gogym/internal/tenant"
	"github.com/gin-gonic/gin"
)

const bearerPrefix = "Bearer "

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
