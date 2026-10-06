package plan

import "github.com/gin-gonic/gin"

func RegisterRoutes(protected *gin.RouterGroup, h *Handler) {
	plans := protected.Group("/athletes/:" + paramAthleteID + "/plans")
	plans.GET("", h.List)
}
