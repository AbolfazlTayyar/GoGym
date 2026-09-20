// Package coach is the one login this app has: signup, login, the JWT that
// gates every other endpoint, and the tenant auth middleware.
package coach

import (
	"time"

	"github.com/google/uuid"
)

// Coach is a coach account — the only login in v1 (docs/spec.md).
type Coach struct {
	ID           uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	FirstName    string    `gorm:"column:first_name;type:text;not null"`
	LastName     string    `gorm:"column:last_name;type:text;not null"`
	Phone        string    `gorm:"column:phone;type:text;not null;unique"`
	PasswordHash string    `gorm:"column:password_hash;type:text;not null"`
	CreatedAt    time.Time `gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt    time.Time `gorm:"column:updated_at;type:timestamptz;not null"`
}

func (Coach) TableName() string {
	return "coach"
}
