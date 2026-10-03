package measurement

import "github.com/gin-gonic/gin"

func RegisterRoutes(protected *gin.RouterGroup, h *Handler) {
	measurements := protected.Group("/athletes/:" + paramAthleteID + "/measurements")
	measurements.POST("", h.Create)
	measurements.GET("", h.List)
}
