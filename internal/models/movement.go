package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Movement.CoachID is NULL for system-seeded movements shared by every coach.
// Deletes are soft so existing block_movement references stay valid.
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
