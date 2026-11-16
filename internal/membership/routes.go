package membership

import "github.com/gin-gonic/gin"

func RegisterRoutes(protected *gin.RouterGroup, h *Handler) {
	protected.GET("/attendance/current", h.Current)

	// A visit is addressed by its plan and date, its primary key; it has no id of its own.
	attendance := protected.Group("/plans/:" + paramPlanID + "/attendance")
	attendance.GET("", h.Get)
	attendance.PUT("", h.Track)
	attendance.DELETE("", h.Untrack)
	attendance.POST("/visits", h.AddVisit)
	attendance.DELETE("/visits/:"+paramDate, h.DeleteVisit)
}
