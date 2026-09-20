package coach

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrNotFound is returned by repository lookups that find no matching row.
var ErrNotFound = errors.New("coach: not found")

// ErrPhoneTaken is returned by Create when the phone number is already
// registered to another coach.
var ErrPhoneTaken = errors.New("coach: phone already registered")

// Repository is the coach table's persistence layer.
//
// Coach is the tenant root, not a tenant-owned table — it has no coach_id
// column to scope by, so it intentionally does not go through
// internal/tenant.Scope. Every other module's repository, once it exists,
// must scope its tenant-owned queries through that helper instead of
// hand-writing WHERE coach_id = ?.
type Repository struct {
	db *gorm.DB
}

// NewRepository builds a Repository backed by db.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create inserts a new coach. It returns ErrPhoneTaken if the phone number
// is already registered.
func (r *Repository) Create(ctx context.Context, c *Coach) error {
	if err := r.db.WithContext(ctx).Create(c).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrPhoneTaken
		}
		return fmt.Errorf("coach: failed to create: %w", err)
	}
	return nil
}

// FindByPhone returns the coach registered under phone, or ErrNotFound.
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

// FindByID returns the coach with the given id, or ErrNotFound.
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
