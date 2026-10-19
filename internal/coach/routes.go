package coach

import "github.com/gin-gonic/gin"

// There is no signup route: coaches are created with cmd/coachctl until a second coach needs self-service signup.
func RegisterRoutes(v1 *gin.RouterGroup, protected *gin.RouterGroup, h *Handler) {
	auth := v1.Group("/auth")
	auth.POST("/login", h.Login)
	auth.POST("/logout", h.Logout)

	protected.GET("/coaches/me", h.Me)
}
