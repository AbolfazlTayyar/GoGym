package plan

import "github.com/gin-gonic/gin"

func RegisterRoutes(protected *gin.RouterGroup, h *Handler) {
	plans := protected.Group("/athletes/:" + paramAthleteID + "/plans")
	plans.GET("", h.List)

	protected.POST("/plans", h.Create)
	protected.GET("/plans/:"+paramPlanID, h.Get)
	protected.POST("/plans/:"+paramPlanID+"/days", h.AddDay)
	protected.POST("/days/:"+paramDayID+"/blocks", h.AddBlock)
	protected.POST("/blocks/:"+paramBlockID+"/movements", h.AddMovements)
}
