package athlete

import (
	"context"
	"fmt"

	"github.com/AbolfazlTayyar/gogym/internal/tenant"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ESCAPE '\' is Postgres's default, stated explicitly because searchPattern escapes with it.
const nameSearchClause = `(first_name ILIKE @pattern ESCAPE '\' OR last_name ILIKE @pattern ESCAPE '\')`

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

	if pattern := searchPattern(opts.Query); pattern != "" {
		query = query.Where(nameSearchClause, map[string]any{"pattern": pattern})
	}

	return query
}
