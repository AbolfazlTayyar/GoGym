package models

import (
	"time"

	"github.com/google/uuid"
)

// Plan will move to internal/plan/model.go once that module exists.
type Plan struct {
	ID        uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	AthleteID uuid.UUID `gorm:"column:athlete_id;type:uuid;not null"`
	StartDate time.Time `gorm:"column:start_date;type:date;not null"`
	Title     string    `gorm:"column:title;type:text;not null"`
	Note      *string   `gorm:"column:note;type:text"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamptz;not null"`
}

func (Plan) TableName() string {
	return "plan"
}
