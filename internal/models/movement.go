package models

import (
	"time"

	"github.com/google/uuid"
)

// Movement will move to internal/movement/model.go once that module exists.
//
// CoachID is nullable: NULL marks a universal, system-seeded movement
// visible to every coach; a non-null value is a coach's own custom movement.
type Movement struct {
	ID          uuid.UUID  `gorm:"column:id;type:uuid;primaryKey"`
	CoachID     *uuid.UUID `gorm:"column:coach_id;type:uuid"`
	Name        string     `gorm:"column:name;type:text;not null"`
	Category    *string    `gorm:"column:category;type:text"`
	Description *string    `gorm:"column:description;type:text"`
	CreatedAt   time.Time  `gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;type:timestamptz;not null"`
}

func (Movement) TableName() string {
	return "movement"
}
