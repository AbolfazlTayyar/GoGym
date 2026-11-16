// Package membership counts the visits of athletes who pay per 12 sessions, so the coach knows when they renew.
package membership

import (
	"time"

	"github.com/google/uuid"
)

// SessionsPerPlan is what every gym the coaches work with sells to an athlete who comes every other day.
const SessionsPerPlan = 12

// SessionPlan marks a plan as tracked. It has no coach_id: ownership is reached through its plan's athlete.
type SessionPlan struct {
	PlanID    uuid.UUID `gorm:"column:plan_id;type:uuid;primaryKey"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz;not null"`
}

func (SessionPlan) TableName() string {
	return "session_plan"
}

type Visit struct {
	PlanID    uuid.UUID `gorm:"column:plan_id;type:uuid;primaryKey"`
	Date      time.Time `gorm:"column:date;type:date;primaryKey"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz;not null"`
}

func (Visit) TableName() string {
	return "attendance"
}

// Attendance is one plan's tracking state: whether it's tracked, and its visits oldest first.
type Attendance struct {
	Tracked bool
	Dates   []time.Time
}

// PlanCount is a tracked current plan's progress, for the athlete rows.
type PlanCount struct {
	PlanID    uuid.UUID
	AthleteID uuid.UUID
	Visits    int
	// LastDate is nil before the first visit.
	LastDate *time.Time
}
