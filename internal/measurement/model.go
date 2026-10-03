// Package measurement records an athlete's body measurements over time, the data behind the progress chart.
package measurement

import (
	"time"

	"github.com/google/uuid"
)

// Measurement has no coach_id: ownership is reached through its athlete, so it can't use tenant.Scope directly.
type Measurement struct {
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

func (Measurement) TableName() string {
	return "athlete_measurement"
}
