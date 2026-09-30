// Package tenant is the single enforcement point for scoping queries to a coach.
package tenant

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const contextKey = "tenant_coach_id"

// SetCoachID must only be called after the request's JWT has been validated.
func SetCoachID(c *gin.Context, coachID uuid.UUID) {
	c.Set(contextKey, coachID)
}

func CoachIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get(contextKey)
	if !ok {
		return uuid.UUID{}, false
	}

	coachID, ok := v.(uuid.UUID)
	return coachID, ok
}

// Scope must back every query on a coach_id table; never hand-write the coach_id filter.
func Scope(db *gorm.DB, coachID uuid.UUID) *gorm.DB {
	return db.Where("coach_id = ?", coachID)
}
