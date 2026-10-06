package plan

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrNotFound also covers a plan whose athlete belongs to another coach.
var ErrNotFound = errors.New("plan: not found")

// Repository trusts the athlete and plan ids it is given; Service must confirm the athlete is the coach's first.
type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// ListByAthlete orders newest first; among plans starting the same day, the later-created one leads.
func (r *Repository) ListByAthlete(ctx context.Context, athleteID uuid.UUID) ([]Plan, error) {
	plans := make([]Plan, 0)
	if err := r.db.WithContext(ctx).
		Where("athlete_id = ?", athleteID).
		Order("start_date DESC, created_at DESC, id DESC").
		Find(&plans).Error; err != nil {
		return nil, fmt.Errorf("plan: failed to list: %w", err)
	}
	return plans, nil
}

// FindByID returns the plan alone; Service checks its athlete before LoadDays touches the children.
func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (*Plan, error) {
	var p Plan
	if err := r.db.WithContext(ctx).First(&p, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("plan: failed to find by id: %w", err)
	}
	return &p, nil
}

// LoadDays issues one query per level (day, block, block_movement, movement), however big the plan is.
func (r *Repository) LoadDays(ctx context.Context, planID uuid.UUID) ([]Day, error) {
	days := make([]Day, 0)
	if err := r.db.WithContext(ctx).
		Where("plan_id = ?", planID).
		Order("order_index, id").
		Preload("Blocks", func(db *gorm.DB) *gorm.DB {
			return db.Order("order_index, created_at, id")
		}).
		Preload("Blocks.Movements", func(db *gorm.DB) *gorm.DB {
			return db.Order("order_in_block, created_at, id")
		}).
		// Unscoped: a soft-deleted movement still has to name itself in the plans that used it.
		Preload("Blocks.Movements.Movement", func(db *gorm.DB) *gorm.DB {
			return db.Unscoped()
		}).
		Find(&days).Error; err != nil {
		return nil, fmt.Errorf("plan: failed to load days: %w", err)
	}
	return days, nil
}
