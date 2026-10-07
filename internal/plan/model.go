// Package plan holds an athlete's training plans and the days, blocks and movements inside them.
package plan

import (
	"time"

	"github.com/AbolfazlTayyar/gogym/internal/models"
	"github.com/google/uuid"
)

// MaxDaysPerPlan is enforced by the day table's unique, 0-6 bounded order_index, not by a count.
const MaxDaysPerPlan = 7

// Plan has no coach_id: ownership is reached through its athlete, so it can't use tenant.Scope directly.
type Plan struct {
	ID        uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	AthleteID uuid.UUID `gorm:"column:athlete_id;type:uuid;not null"`
	StartDate time.Time `gorm:"column:start_date;type:date;not null"`
	Title     string    `gorm:"column:title;type:text;not null"`
	Note      *string   `gorm:"column:note;type:text"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamptz;not null"`
	Days      []Day     `gorm:"foreignKey:PlanID"`
}

func (Plan) TableName() string {
	return "plan"
}

// Day, Block and BlockMovement carry created_at only, because the builder deletes and recreates
// them instead of editing in place. That is safe only while nothing outside this tree holds their ids.
//
// Workout logging must not break that. A log row that foreign-keys to block_movement (or block, or
// day) is orphaned by every plan edit, and training history is the hardest data in the product to
// replace. A log references movement_id plus a snapshot of what was prescribed at the time (sets,
// reps, duration and load as programmed), never the plan structure.
//
// Plan templates and duplication pull the other way: they copy a plan and then edit the copy in
// place. If they are coming, delete-and-recreate is on borrowed time, and updated_at on these three
// tables is cheap to add while there is no data to backfill.

type Day struct {
	ID         uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	PlanID     uuid.UUID `gorm:"column:plan_id;type:uuid;not null"`
	Label      string    `gorm:"column:label;type:text;not null"`
	OrderIndex int       `gorm:"column:order_index;type:int;not null"`
	CreatedAt  time.Time `gorm:"column:created_at;type:timestamptz;not null"`
	Blocks     []Block   `gorm:"foreignKey:DayID"`
}

func (Day) TableName() string {
	return "day"
}

type Block struct {
	ID          uuid.UUID       `gorm:"column:id;type:uuid;primaryKey"`
	DayID       uuid.UUID       `gorm:"column:day_id;type:uuid;not null"`
	OrderIndex  int             `gorm:"column:order_index;type:int;not null"`
	Sets        int             `gorm:"column:sets;type:int;not null"`
	RestSeconds *int            `gorm:"column:rest_seconds;type:int"`
	Notes       *string         `gorm:"column:notes;type:text"`
	CreatedAt   time.Time       `gorm:"column:created_at;type:timestamptz;not null"`
	Movements   []BlockMovement `gorm:"foreignKey:BlockID"`
}

func (Block) TableName() string {
	return "block"
}

// BlockMovement.Load is free text, not kg: coaches prescribe %1RM, RPE, or bodyweight just as often.
type BlockMovement struct {
	ID              uuid.UUID       `gorm:"column:id;type:uuid;primaryKey"`
	BlockID         uuid.UUID       `gorm:"column:block_id;type:uuid;not null"`
	MovementID      uuid.UUID       `gorm:"column:movement_id;type:uuid;not null"`
	Reps            *int            `gorm:"column:reps;type:int"`
	DurationSeconds *int            `gorm:"column:duration_seconds;type:int"`
	Load            *string         `gorm:"column:load;type:text"`
	OrderInBlock    int             `gorm:"column:order_in_block;type:int;not null"`
	CreatedAt       time.Time       `gorm:"column:created_at;type:timestamptz;not null"`
	Movement        models.Movement `gorm:"foreignKey:MovementID"`
}

func (BlockMovement) TableName() string {
	return "block_movement"
}
