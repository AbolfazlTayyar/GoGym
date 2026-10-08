package movement

import (
	"context"
	"errors"
	"fmt"

	"github.com/AbolfazlTayyar/gogym/internal/search"
	"github.com/AbolfazlTayyar/gogym/internal/tenant"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ESCAPE '\' is Postgres's default, stated explicitly because search.ContainsPattern escapes with it.
const nameSearchClause = `name ILIKE ? ESCAPE '\'`

// ErrNotFound also covers another coach's movement, so callers can 404 without revealing it exists.
var ErrNotFound = errors.New("movement: not found")

// Repository reads through tenant.ScopeWithUniversal but writes only through tenant.Scope, so a
// universal movement can be listed and never changed, whatever the caller checked first.
type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, m *Movement) error {
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return fmt.Errorf("movement: failed to create: %w", err)
	}
	return nil
}

// FindVisible also returns a universal movement; the caller decides whether it may be changed.
func (r *Repository) FindVisible(ctx context.Context, coachID, id uuid.UUID) (*Movement, error) {
	var m Movement
	if err := tenant.ScopeWithUniversal(r.db.WithContext(ctx), coachID).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("movement: failed to find by id: %w", err)
	}
	return &m, nil
}

// FindVisibleByIDs leaves out ids the coach can't use: unknown, another coach's, or soft-deleted.
func (r *Repository) FindVisibleByIDs(ctx context.Context, coachID uuid.UUID, ids []uuid.UUID) ([]Movement, error) {
	movements := make([]Movement, 0, len(ids))
	if err := tenant.ScopeWithUniversal(r.db.WithContext(ctx), coachID).
		Where("id IN ?", ids).
		Find(&movements).Error; err != nil {
		return nil, fmt.Errorf("movement: failed to find by ids: %w", err)
	}
	return movements, nil
}

// List expects opts already normalized by normalizeListOptions.
func (r *Repository) List(ctx context.Context, coachID uuid.UUID, opts ListOptions) ([]Movement, error) {
	query := tenant.ScopeWithUniversal(r.db.WithContext(ctx), coachID)

	if pattern := search.ContainsPattern(opts.Query); pattern != "" {
		query = query.Where(nameSearchClause, pattern)
	}
	if opts.MuscleGroup != "" {
		query = query.Where("muscle_group = ?", opts.MuscleGroup)
	}
	if opts.Equipment != "" {
		query = query.Where("equipment = ?", opts.Equipment)
	}

	movements := make([]Movement, 0)
	if err := query.Order("name, id").Find(&movements).Error; err != nil {
		return nil, fmt.Errorf("movement: failed to list: %w", err)
	}
	return movements, nil
}

// Update writes the columns a coach edits and leaves media_url alone; updated_at is set by GORM.
func (r *Repository) Update(ctx context.Context, coachID uuid.UUID, m *Movement) error {
	result := tenant.Scope(r.db.WithContext(ctx), coachID).
		Model(m).
		Select("name", "category", "description", "muscle_group", "equipment").
		Updates(m)
	if result.Error != nil {
		return fmt.Errorf("movement: failed to update: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete is soft, so plans that already use the movement keep naming it.
func (r *Repository) Delete(ctx context.Context, coachID, id uuid.UUID) error {
	result := tenant.Scope(r.db.WithContext(ctx), coachID).Delete(&Movement{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("movement: failed to delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
