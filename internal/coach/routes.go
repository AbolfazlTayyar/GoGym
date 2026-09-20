package coach

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts /auth/signup and /auth/login (unauthenticated, rate
// limited) and /coaches/me (behind AuthMiddleware) onto v1.
func RegisterRoutes(v1 *gin.RouterGroup, protected *gin.RouterGroup, h *Handler) {
	auth := v1.Group("/auth")
	auth.POST("/signup", h.Signup)
	auth.POST("/login", h.Login)

	protected.GET("/coaches/me", h.Me)
}
