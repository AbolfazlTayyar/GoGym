package measurement

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrNotFound is an unknown measurement id. One whose athlete belongs to another coach is athlete.ErrNotFound
// instead; both answer 404.
var ErrNotFound = errors.New("measurement: not found")

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

// FindByID checks no ownership: the caller must check the measurement's athlete before using it.
func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (*Measurement, error) {
	var m Measurement
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("measurement: failed to find by id: %w", err)
	}
	return &m, nil
}

// Update writes every column a coach edits, a cleared metric as NULL; updated_at is set by GORM.
func (r *Repository) Update(ctx context.Context, m *Measurement) error {
	result := r.db.WithContext(ctx).
		Model(m).
		Select("date", "weight", "chest", "waist", "arm", "thigh", "hip").
		Updates(m)
	if result.Error != nil {
		return fmt.Errorf("measurement: failed to update: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, m *Measurement) error {
	result := r.db.WithContext(ctx).Delete(m)
	if result.Error != nil {
		return fmt.Errorf("measurement: failed to delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
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
