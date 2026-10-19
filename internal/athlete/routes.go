package athlete

import "github.com/gin-gonic/gin"

func RegisterRoutes(protected *gin.RouterGroup, h *Handler) {
	athletes := protected.Group("/athletes")
	athletes.POST("", h.Create)
	athletes.GET("", h.List)
	athletes.GET("/:"+paramID, h.Get)
	athletes.PUT("/:"+paramID, h.Update)
	athletes.DELETE("/:"+paramID, h.Delete)
}
