package membership

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrNotTracked    = errors.New("membership: plan is not tracked")
	ErrVisitNotFound = errors.New("membership: visit not found")
	ErrDuplicateDate = errors.New("membership: date already recorded")
	ErrSessionsUsed  = errors.New("membership: every session is used")
)

// Repository trusts the plan id it is given; Service must confirm the plan is the coach's before calling it.
type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Find(ctx context.Context, planID uuid.UUID) (*Attendance, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&SessionPlan{}).Where("plan_id = ?", planID).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("membership: failed to find session plan: %w", err)
	}

	a := &Attendance{Tracked: count > 0, Dates: make([]time.Time, 0)}
	if !a.Tracked {
		return a, nil
	}

	if err := r.db.WithContext(ctx).Model(&Visit{}).
		Where("plan_id = ?", planID).
		Order("date ASC").
		Pluck("date", &a.Dates).Error; err != nil {
		return nil, fmt.Errorf("membership: failed to list visits: %w", err)
	}

	return a, nil
}

// Track is idempotent: turning on a tracked plan keeps its visits.
func (r *Repository) Track(ctx context.Context, planID uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&SessionPlan{PlanID: planID}).Error; err != nil {
		return fmt.Errorf("membership: failed to track: %w", err)
	}
	return nil
}

// Untrack deletes the plan's visits with it, through the attendance foreign key's cascade.
func (r *Repository) Untrack(ctx context.Context, planID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&SessionPlan{}, "plan_id = ?", planID).Error; err != nil {
		return fmt.Errorf("membership: failed to untrack: %w", err)
	}
	return nil
}

// AddVisit locks the plan's session_plan row, so two taps at once can't both pass the count and make a 13th visit.
func (r *Repository) AddVisit(ctx context.Context, planID uuid.UUID, date time.Time) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var sp SessionPlan
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&sp, "plan_id = ?", planID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotTracked
			}
			return fmt.Errorf("membership: failed to lock session plan: %w", err)
		}

		var count int64
		if err := tx.Model(&Visit{}).Where("plan_id = ?", planID).Count(&count).Error; err != nil {
			return fmt.Errorf("membership: failed to count visits: %w", err)
		}
		if count >= SessionsPerPlan {
			return ErrSessionsUsed
		}

		if err := tx.Create(&Visit{PlanID: planID, Date: date}).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return ErrDuplicateDate
			}
			return fmt.Errorf("membership: failed to add visit: %w", err)
		}
		return nil
	})
}

func (r *Repository) DeleteVisit(ctx context.Context, planID uuid.UUID, date time.Time) error {
	result := r.db.WithContext(ctx).Delete(&Visit{}, "plan_id = ? AND date = ?", planID, date)
	if result.Error != nil {
		return fmt.Errorf("membership: failed to delete visit: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrVisitNotFound
	}
	return nil
}

// Counts returns a row for each of planIDs that is tracked, untracked ones left out.
func (r *Repository) Counts(ctx context.Context, planIDs []uuid.UUID) ([]PlanCount, error) {
	counts := make([]PlanCount, 0)
	if len(planIDs) == 0 {
		return counts, nil
	}

	if err := r.db.WithContext(ctx).
		Table("session_plan").
		Select("session_plan.plan_id, COUNT(attendance.date) AS visits, MAX(attendance.date) AS last_date").
		Joins("LEFT JOIN attendance ON attendance.plan_id = session_plan.plan_id").
		Where("session_plan.plan_id IN ?", planIDs).
		Group("session_plan.plan_id").
		Scan(&counts).Error; err != nil {
		return nil, fmt.Errorf("membership: failed to count visits: %w", err)
	}

	return counts, nil
}
