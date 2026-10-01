// Package athlete manages a coach's athlete roster.
package athlete

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Athlete types are the coach's service tier for an athlete, not a training format.
const (
	// TypePrivate is managed on an ongoing basis; it is the column default and the dashboard's list.
	TypePrivate = "private"
	// TypePublic is a one-off plan-link delivery, kept off the dashboard unless asked for.
	TypePublic = "public"
)

// Experience levels have no DB constraint; the service is the only thing enforcing this set.
const (
	ExperienceBeginner     = "beginner"
	ExperienceIntermediate = "intermediate"
	ExperienceAdvanced     = "advanced"
)

type Athlete struct {
	ID              uuid.UUID      `gorm:"column:id;type:uuid;primaryKey"`
	CoachID         uuid.UUID      `gorm:"column:coach_id;type:uuid;not null"`
	FirstName       string         `gorm:"column:first_name;type:text;not null"`
	LastName        string         `gorm:"column:last_name;type:text;not null"`
	Phone           string         `gorm:"column:phone;type:text;not null"`
	ExperienceLevel *string        `gorm:"column:experience_level;type:text"`
	Injuries        *string        `gorm:"column:injuries;type:text"`
	Goal            *string        `gorm:"column:goal;type:text"`
	Height          *float64       `gorm:"column:height;type:numeric"`
	AthleteType     string         `gorm:"column:athlete_type;type:text;not null;default:private"`
	CreatedAt       time.Time      `gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt       time.Time      `gorm:"column:updated_at;type:timestamptz;not null"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz;index"`
}

func (Athlete) TableName() string {
	return "athlete"
}
