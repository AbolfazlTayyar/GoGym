package coach

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("coach: not found")

var ErrPhoneTaken = errors.New("coach: phone already registered")

// Repository skips tenant.Scope on purpose: coach is the tenant root and has no coach_id.
type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, c *Coach) error {
	if err := r.db.WithContext(ctx).Create(c).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrPhoneTaken
		}
		return fmt.Errorf("coach: failed to create: %w", err)
	}
	return nil
}

func (r *Repository) FindByPhone(ctx context.Context, phone string) (*Coach, error) {
	var c Coach
	if err := r.db.WithContext(ctx).First(&c, "phone = ?", phone).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("coach: failed to find by phone: %w", err)
	}
	return &c, nil
}

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (*Coach, error) {
	var c Coach
	if err := r.db.WithContext(ctx).First(&c, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("coach: failed to find by id: %w", err)
	}
	return &c, nil
}
