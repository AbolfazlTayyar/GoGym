package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Movement will move to internal/movement/model.go once that module exists.
//
// CoachID is nullable: NULL marks a universal, system-seeded movement
// visible to every coach; a non-null value is a coach's own custom movement.
//
// DeletedAt is GORM soft delete: coaches can delete their own movements from
// the UI, and this keeps them out of query results (via GORM's default
// scope) without breaking existing block_movement references. The coach's
// effective library query (coach_id = :coach_id OR coach_id IS NULL) is
// unaffected — the soft-delete scope simply ANDs deleted_at IS NULL onto it.
type Movement struct {
	ID          uuid.UUID      `gorm:"column:id;type:uuid;primaryKey"`
	CoachID     *uuid.UUID     `gorm:"column:coach_id;type:uuid"`
	Name        string         `gorm:"column:name;type:text;not null"`
	Category    *string        `gorm:"column:category;type:text"`
	Description *string        `gorm:"column:description;type:text"`
	MediaURL    *string        `gorm:"column:media_url;type:text"`
	CreatedAt   time.Time      `gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;type:timestamptz;not null"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz;index"`
}

func (Movement) TableName() string {
	return "movement"
}
