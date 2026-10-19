package plan

import "github.com/gin-gonic/gin"

func RegisterRoutes(protected *gin.RouterGroup, h *Handler) {
	plans := protected.Group("/athletes/:" + paramAthleteID + "/plans")
	plans.GET("", h.List)

	protected.POST("/plans", h.Create)
	// Gin matches the static segment before the :id wildcard, so "current" is never read as a plan id.
	protected.GET("/plans/current", h.Current)
	protected.GET("/plans/:"+paramPlanID, h.Get)
	protected.PUT("/plans/:"+paramPlanID, h.Update)
	protected.DELETE("/plans/:"+paramPlanID, h.Delete)
	protected.POST("/plans/:"+paramPlanID+"/days", h.AddDay)
	protected.POST("/days/:"+paramDayID+"/blocks", h.AddBlock)
	protected.PUT("/days/:"+paramDayID+"/blocks", h.SaveDay)
	protected.DELETE("/days/:"+paramDayID, h.DeleteDay)
	protected.POST("/blocks/:"+paramBlockID+"/movements", h.AddMovements)
}
