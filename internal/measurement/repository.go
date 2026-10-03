package measurement

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository trusts the athleteID it is given; Service must confirm the athlete is the coach's before calling it.
type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, m *Measurement) error {
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return fmt.Errorf("measurement: failed to create: %w", err)
	}
	return nil
}

// ListByAthlete orders oldest first, the order a trend chart plots in; ties keep insertion order.
func (r *Repository) ListByAthlete(ctx context.Context, athleteID uuid.UUID) ([]Measurement, error) {
	measurements := make([]Measurement, 0)
	if err := r.db.WithContext(ctx).
		Where("athlete_id = ?", athleteID).
		Order("date ASC, created_at ASC, id ASC").
		Find(&measurements).Error; err != nil {
		return nil, fmt.Errorf("measurement: failed to list: %w", err)
	}
	return measurements, nil
}
