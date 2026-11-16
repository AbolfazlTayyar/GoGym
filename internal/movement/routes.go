package movement

import "github.com/gin-gonic/gin"

func RegisterRoutes(protected *gin.RouterGroup, h *Handler) {
	movements := protected.Group("/movements")
	movements.GET("", h.List)
	movements.POST("", h.Create)
	movements.PUT("/:"+paramID, h.Update)
	movements.DELETE("/:"+paramID, h.Delete)

	// A catch-all, because a category name may hold a "/" that gin has already unescaped.
	categories := protected.Group("/movement-categories")
	categories.PUT("/*"+paramCategory, h.RenameCategory)
	categories.DELETE("/*"+paramCategory, h.DeleteCategory)
}
