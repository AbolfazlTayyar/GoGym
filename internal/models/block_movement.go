package models

import (
	"time"

	"github.com/google/uuid"
)

// BlockMovement will move to internal/plan/model.go once that module
// exists. It is the join between a Block and the shared Movement library.
type BlockMovement struct {
	ID              uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	BlockID         uuid.UUID `gorm:"column:block_id;type:uuid;not null"`
	MovementID      uuid.UUID `gorm:"column:movement_id;type:uuid;not null"`
	Reps            *int      `gorm:"column:reps;type:int"`
	DurationSeconds *int      `gorm:"column:duration_seconds;type:int"`
	OrderInBlock    int       `gorm:"column:order_in_block;type:int;not null"`
	CreatedAt       time.Time `gorm:"column:created_at;type:timestamptz;not null"`
}

func (BlockMovement) TableName() string {
	return "block_movement"
}
