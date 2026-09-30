package models

import (
	"time"

	"github.com/google/uuid"
)

type Block struct {
	ID          uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	DayID       uuid.UUID `gorm:"column:day_id;type:uuid;not null"`
	OrderIndex  int       `gorm:"column:order_index;type:int;not null"`
	Sets        int       `gorm:"column:sets;type:int;not null"`
	RestSeconds *int      `gorm:"column:rest_seconds;type:int"`
	Notes       *string   `gorm:"column:notes;type:text"`
	CreatedAt   time.Time `gorm:"column:created_at;type:timestamptz;not null"`
}

func (Block) TableName() string {
	return "block"
}
