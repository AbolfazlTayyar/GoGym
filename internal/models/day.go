package models

import (
	"time"

	"github.com/google/uuid"
)

type Day struct {
	ID         uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	PlanID     uuid.UUID `gorm:"column:plan_id;type:uuid;not null"`
	Label      string    `gorm:"column:label;type:text;not null"`
	OrderIndex int       `gorm:"column:order_index;type:int;not null"`
	CreatedAt  time.Time `gorm:"column:created_at;type:timestamptz;not null"`
}

func (Day) TableName() string {
	return "day"
}
