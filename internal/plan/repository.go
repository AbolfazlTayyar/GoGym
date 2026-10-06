package plan

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
