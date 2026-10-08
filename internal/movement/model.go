// Package movement is a coach's exercise library: their own movements plus the universal, system-seeded ones.
package movement

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// muscleGroups and equipmentTypes mirror the movement table's CHECK constraints: a new value needs a
// migration as well as an entry here.
var (
	muscleGroups   = []string{"chest", "back", "shoulders", "arms", "legs", "core", "full_body"}
	equipmentTypes = []string{"bodyweight", "barbell", "dumbbell", "kettlebell", "machine", "cable", "band", "other"}
)

// Movement.CoachID is NULL for a universal movement, which every coach can read and none can change.
// Deletes are soft so existing block_movement references stay valid.
type Movement struct {
	ID          uuid.UUID      `gorm:"column:id;type:uuid;primaryKey"`
	CoachID     *uuid.UUID     `gorm:"column:coach_id;type:uuid"`
	Name        string         `gorm:"column:name;type:text;not null"`
	Category    *string        `gorm:"column:category;type:text"`
	Description *string        `gorm:"column:description;type:text"`
	MuscleGroup *string        `gorm:"column:muscle_group;type:text"`
	Equipment   *string        `gorm:"column:equipment;type:text"`
	MediaURL    *string        `gorm:"column:media_url;type:text"`
	CreatedAt   time.Time      `gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;type:timestamptz;not null"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz;index"`
}

func (Movement) TableName() string {
	return "movement"
}

// IsUniversal reports a system-seeded movement, which no coach owns.
func (m *Movement) IsUniversal() bool {
	return m.CoachID == nil
}
