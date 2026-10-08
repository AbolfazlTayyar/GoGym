package movement

import "github.com/gin-gonic/gin"

func RegisterRoutes(protected *gin.RouterGroup, h *Handler) {
	movements := protected.Group("/movements")
	movements.GET("", h.List)
	movements.POST("", h.Create)
	movements.PUT("/:"+paramID, h.Update)
	movements.DELETE("/:"+paramID, h.Delete)
}
