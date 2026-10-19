package plan

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrNotFound also covers a plan, day or block whose athlete belongs to another coach.
var ErrNotFound = errors.New("plan: not found")

// day has one CHECK and one UNIQUE besides its key, so the portable sentinels map to these unambiguously.
var (
	ErrDaySlotOutOfRange = errors.New("plan: day order_index outside the plan's slots")
	ErrDaySlotTaken      = errors.New("plan: day order_index already taken")
)

// Repository trusts the athlete, plan, day and block ids it is given; Service must confirm the athlete is the coach's first.
type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// newestFirst is the order currentPlanID relies on: among plans starting the same day, the later-created one leads.
const newestFirst = "start_date DESC, created_at DESC, id DESC"

func (r *Repository) ListByAthlete(ctx context.Context, athleteID uuid.UUID) ([]Plan, error) {
	plans := make([]Plan, 0)
	if err := r.db.WithContext(ctx).
		Where("athlete_id = ?", athleteID).
		Order(newestFirst).
		Find(&plans).Error; err != nil {
		return nil, fmt.Errorf("plan: failed to list: %w", err)
	}
	return plans, nil
}

// ListByAthletes loads every given athlete's plans in one query, each athlete's newest first.
func (r *Repository) ListByAthletes(ctx context.Context, athleteIDs []uuid.UUID) ([]Plan, error) {
	plans := make([]Plan, 0)
	if len(athleteIDs) == 0 {
		return plans, nil
	}
	if err := r.db.WithContext(ctx).
		Where("athlete_id IN ?", athleteIDs).
		Order("athlete_id, " + newestFirst).
		Find(&plans).Error; err != nil {
		return nil, fmt.Errorf("plan: failed to list for athletes: %w", err)
	}
	return plans, nil
}

// Size is what a plan summary says about a plan's contents.
type Size struct {
	// DayLabels are in order_index order; an empty plan has none, never nil.
	DayLabels []string
	// BlockCount counts numbered lines: a superset is one block, however many movements it holds.
	BlockCount int
}

// Sizes loads every given plan's day labels and block count in two queries, however many plans there are.
func (r *Repository) Sizes(ctx context.Context, planIDs []uuid.UUID) (map[uuid.UUID]Size, error) {
	sizes := make(map[uuid.UUID]Size, len(planIDs))
	for _, id := range planIDs {
		sizes[id] = Size{DayLabels: []string{}}
	}
	if len(planIDs) == 0 {
		return sizes, nil
	}

	var days []Day
	if err := r.db.WithContext(ctx).
		Select("plan_id", "label").
		Where("plan_id IN ?", planIDs).
		Order("plan_id, order_index").
		Find(&days).Error; err != nil {
		return nil, fmt.Errorf("plan: failed to load day labels: %w", err)
	}
	for _, d := range days {
		size := sizes[d.PlanID]
		size.DayLabels = append(size.DayLabels, d.Label)
		sizes[d.PlanID] = size
	}

	var counts []struct {
		PlanID uuid.UUID
		Blocks int
	}
	if err := r.db.WithContext(ctx).
		Model(&Block{}).
		Select("day.plan_id AS plan_id, COUNT(*) AS blocks").
		Joins("JOIN day ON day.id = block.day_id").
		Where("day.plan_id IN ?", planIDs).
		Group("day.plan_id").
		Scan(&counts).Error; err != nil {
		return nil, fmt.Errorf("plan: failed to count blocks: %w", err)
	}
	for _, c := range counts {
		size := sizes[c.PlanID]
		size.BlockCount = c.Blocks
		sizes[c.PlanID] = size
	}

	return sizes, nil
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

// AthleteIDOfDay only resolves which athlete the day hangs from; it says nothing about that athlete's coach.
func (r *Repository) AthleteIDOfDay(ctx context.Context, dayID uuid.UUID) (uuid.UUID, error) {
	return athleteIDOf(r.db.WithContext(ctx).
		Model(&Day{}).
		Joins("JOIN plan ON plan.id = day.plan_id").
		Where("day.id = ?", dayID))
}

// AthleteIDOfBlock only resolves which athlete the block hangs from; it says nothing about that athlete's coach.
func (r *Repository) AthleteIDOfBlock(ctx context.Context, blockID uuid.UUID) (uuid.UUID, error) {
	return athleteIDOf(r.db.WithContext(ctx).
		Model(&Block{}).
		Joins("JOIN day ON day.id = block.day_id").
		Joins("JOIN plan ON plan.id = day.plan_id").
		Where("block.id = ?", blockID))
}

func athleteIDOf(query *gorm.DB) (uuid.UUID, error) {
	var ids []uuid.UUID
	if err := query.Pluck("plan.athlete_id", &ids).Error; err != nil {
		return uuid.Nil, fmt.Errorf("plan: failed to resolve athlete: %w", err)
	}
	if len(ids) == 0 {
		return uuid.Nil, ErrNotFound
	}
	return ids[0], nil
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

func (r *Repository) Create(ctx context.Context, p *Plan) error {
	if err := r.db.WithContext(ctx).Omit(clause.Associations).Create(p).Error; err != nil {
		return fmt.Errorf("plan: failed to create: %w", err)
	}
	return nil
}

func (r *Repository) Update(ctx context.Context, p *Plan) error {
	result := r.db.WithContext(ctx).
		Model(p).
		Select("title", "note", "start_date", "updated_at").
		Updates(p)
	if result.Error != nil {
		return fmt.Errorf("plan: failed to update: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete is hard: the schema's ON DELETE CASCADE removes the plan's days, blocks and block movements.
func (r *Repository) Delete(ctx context.Context, planID uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&Plan{}, "id = ?", planID)
	if result.Error != nil {
		return fmt.Errorf("plan: failed to delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// CreateDay leaves the 7-day cap to the schema, which holds it even against concurrent inserts.
func (r *Repository) CreateDay(ctx context.Context, d *Day) error {
	err := r.db.WithContext(ctx).Omit(clause.Associations).Create(d).Error
	switch {
	case err == nil:
		return nil
	case errors.Is(err, gorm.ErrCheckConstraintViolated):
		return ErrDaySlotOutOfRange
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return ErrDaySlotTaken
	default:
		return fmt.Errorf("plan: failed to create day: %w", err)
	}
}

func (r *Repository) CreateBlock(ctx context.Context, b *Block) error {
	if err := r.db.WithContext(ctx).Omit(clause.Associations).Create(b).Error; err != nil {
		return fmt.Errorf("plan: failed to create block: %w", err)
	}
	return nil
}

// CreateBlockMovements inserts all or none; the explicit transaction keeps that true whatever GORM's
// default-transaction and batch-size settings later become.
func (r *Repository) CreateBlockMovements(ctx context.Context, movements []BlockMovement) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Omit(clause.Associations).Create(&movements).Error
	})
	if err != nil {
		return fmt.Errorf("plan: failed to create block movements: %w", err)
	}
	return nil
}
