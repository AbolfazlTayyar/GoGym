package athlete

import (
	"context"
	"errors"
	"fmt"

	"github.com/AbolfazlTayyar/gogym/internal/platform/search"
	"github.com/AbolfazlTayyar/gogym/internal/platform/tenant"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ESCAPE '\' is Postgres's default, stated explicitly because search.ContainsPattern escapes with it.
const nameSearchClause = `(first_name ILIKE @pattern ESCAPE '\' OR last_name ILIKE @pattern ESCAPE '\')`

// ErrNotFound also covers another coach's athlete, so callers can 404 without revealing it exists.
var ErrNotFound = errors.New("athlete: not found")

// Repository queries athlete, a tenant-owned table, so every query must start from tenant.Scope.
type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, a *Athlete) error {
	if err := r.db.WithContext(ctx).Create(a).Error; err != nil {
		return fmt.Errorf("athlete: failed to create: %w", err)
	}
	return nil
}

func (r *Repository) FindByID(ctx context.Context, coachID, id uuid.UUID) (*Athlete, error) {
	var a Athlete
	if err := tenant.Scope(r.db.WithContext(ctx), coachID).First(&a, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("athlete: failed to find by id: %w", err)
	}
	return &a, nil
}

// editableColumns are the ones Update writes, nil values included, so a cleared optional field becomes NULL.
var editableColumns = []string{
	"first_name", "last_name", "phone", "experience_level", "injuries", "goal", "height", "athlete_type", "updated_at",
}

func (r *Repository) Update(ctx context.Context, coachID uuid.UUID, a *Athlete) error {
	result := tenant.Scope(r.db.WithContext(ctx), coachID).
		Model(a).
		Select(editableColumns).
		Updates(a)
	if result.Error != nil {
		return fmt.Errorf("athlete: failed to update: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete is soft: deleted_at hides the athlete from every tenant-scoped query.
func (r *Repository) Delete(ctx context.Context, coachID, id uuid.UUID) error {
	result := tenant.Scope(r.db.WithContext(ctx), coachID).Delete(&Athlete{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("athlete: failed to delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// IDs returns every one of the coach's athletes, private and public, newest first, unpaginated.
func (r *Repository) IDs(ctx context.Context, coachID uuid.UUID) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0)
	if err := tenant.Scope(r.db.WithContext(ctx), coachID).
		Model(&Athlete{}).
		Order("created_at DESC, id DESC").
		Pluck("id", &ids).Error; err != nil {
		return nil, fmt.Errorf("athlete: failed to list ids: %w", err)
	}
	return ids, nil
}

// List expects opts already normalized by normalizeListOptions.
func (r *Repository) List(ctx context.Context, coachID uuid.UUID, opts ListOptions) ([]Athlete, int64, error) {
	query := r.filtered(ctx, coachID, opts)

	var total int64
	if err := query.Model(&Athlete{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("athlete: failed to count: %w", err)
	}

	athletes := make([]Athlete, 0, opts.Limit)
	if err := r.filtered(ctx, coachID, opts).
		Order("created_at DESC, id DESC").
		Limit(opts.Limit).
		Offset(opts.Offset).
		Find(&athletes).Error; err != nil {
		return nil, 0, fmt.Errorf("athlete: failed to list: %w", err)
	}

	return athletes, total, nil
}

// filtered is rebuilt per statement: GORM mutates a chained *gorm.DB, so sharing one leaks Count state into Find.
func (r *Repository) filtered(ctx context.Context, coachID uuid.UUID, opts ListOptions) *gorm.DB {
	query := tenant.Scope(r.db.WithContext(ctx), coachID)

	if opts.Type != TypeFilterAll {
		query = query.Where("athlete_type = ?", opts.Type)
	}

	if pattern := search.ContainsPattern(opts.Query); pattern != "" {
		query = query.Where(nameSearchClause, map[string]any{"pattern": pattern})
	}

	return query
}
