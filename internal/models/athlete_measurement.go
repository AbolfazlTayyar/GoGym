package models

import (
	"time"

	"github.com/google/uuid"
)

// AthleteMeasurement will move to internal/athlete/model.go once that
// module exists.
type AthleteMeasurement struct {
	ID        uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	AthleteID uuid.UUID `gorm:"column:athlete_id;type:uuid;not null"`
	Date      time.Time `gorm:"column:date;type:date;not null"`
	Weight    *float64  `gorm:"column:weight;type:numeric"`
	Chest     *float64  `gorm:"column:chest;type:numeric"`
	Waist     *float64  `gorm:"column:waist;type:numeric"`
	Arm       *float64  `gorm:"column:arm;type:numeric"`
	Thigh     *float64  `gorm:"column:thigh;type:numeric"`
	Hip       *float64  `gorm:"column:hip;type:numeric"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamptz;not null"`
}

func (AthleteMeasurement) TableName() string {
	return "athlete_measurement"
}
