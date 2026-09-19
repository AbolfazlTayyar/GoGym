package models

import (
	"time"

	"github.com/google/uuid"
)

// Athlete will move to internal/athlete/model.go once that module exists.
type Athlete struct {
	ID              uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	CoachID         uuid.UUID `gorm:"column:coach_id;type:uuid;not null"`
	FirstName       string    `gorm:"column:first_name;type:text;not null"`
	LastName        string    `gorm:"column:last_name;type:text;not null"`
	Phone           string    `gorm:"column:phone;type:text;not null"`
	ExperienceLevel *string   `gorm:"column:experience_level;type:text"`
	Injuries        *string   `gorm:"column:injuries;type:text"`
	Goal            *string   `gorm:"column:goal;type:text"`
	Height          *float64  `gorm:"column:height;type:numeric"`
	AthleteType     string    `gorm:"column:athlete_type;type:text;not null;default:private"`
	CreatedAt       time.Time `gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt       time.Time `gorm:"column:updated_at;type:timestamptz;not null"`
}

func (Athlete) TableName() string {
	return "athlete"
}
