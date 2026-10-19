package plan

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/AbolfazlTayyar/gogym/internal/athlete"
	"github.com/AbolfazlTayyar/gogym/internal/movement"
	"github.com/AbolfazlTayyar/gogym/internal/platform/httpx"
	"github.com/google/uuid"
)

// Length limits count runes, not bytes, so Persian text gets the same allowance as Latin.
const (
	maxTitleLength = 100
	maxNotesLength = 2000
	maxLoadLength  = 100
)

// A saved day is bounded so one request can't insert thousands of rows; her days are 6-9 lines with at
// most tri-sets, so these sit well above real use.
const (
	maxBlocksPerDay      = 40
	maxMovementsPerBlock = 10
)

// Numeric bounds catch a unit mix-up such as milliseconds sent as seconds, and keep every value inside
// Postgres INT, which would otherwise reject it as a 500; they don't judge the programming.
const (
	maxSets            = 50
	maxReps            = 1000
	maxRestSeconds     = 60 * 60
	maxDurationSeconds = 4 * 60 * 60
	// maxPosition bounds block and movement ordering keys; sparse keys like 10, 20, 30 still fit.
	maxPosition = 999
)

const (
	fieldOrderIndex   = "order_index"
	fieldMovementID   = "movement_id"
	fieldSourcePlanID = "source_plan_id"
)

const (
	msgUUID         = "must be a UUID"
	msgMaxLength    = "must be at most %d characters"
	msgRange        = "must be between %d and %d"
	msgDayLabel     = "must be one of A-G or day1-day7"
	msgDaySlotTaken = "is already used by another day in this plan"
	// msgNotFound answers an id in the body the coach can't use, without saying whether it exists.
	msgNotFound    = "not found"
	msgTooMany     = "must have at most %d items"
	msgNoMovements = "at least one movement is required"
)

// dayLabels names each of a plan's MaxDaysPerPlan slots, in either the letter or the numbered scheme.
var dayLabels = []string{
	"A", "B", "C", "D", "E", "F", "G",
	"day1", "day2", "day3", "day4", "day5", "day6", "day7",
}

// ValidationError's Fields are keyed by the json names the client sent, as "[i].name" inside an array body.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("plan: validation failed for %d field(s)", len(e.Fields))
}

// ErrNoMovements has no field to point at: the body itself is the empty array.
var ErrNoMovements = errors.New("plan: no movements given")

// ErrTooManyBlocks has no field to point at either: the body is the over-long array.
var ErrTooManyBlocks = fmt.Errorf("plan: a day holds at most %d blocks", maxBlocksPerDay)

// ListResult's CurrentID is uuid.Nil when the athlete has no plan that has started yet.
type ListResult struct {
	Plans     []Plan
	CurrentID uuid.UUID
	Sizes     map[uuid.UUID]Size
}

// CreateInput's ids are strings so a malformed one is reported as a field error like any other.
type CreateInput struct {
	AthleteID string
	StartDate string
	Title     string
	Note      *string
	// SourcePlanID, when set, makes the plan a copy of that plan's days; empty means an empty plan.
	SourcePlanID string
}

// UpdateInput has no athlete: giving a plan to another athlete is a copy, not an edit.
type UpdateInput struct {
	StartDate string
	Title     string
	Note      *string
}

// SaveBlockInput is one line of a saved day; its position in the request is its order.
type SaveBlockInput struct {
	Sets        *int
	RestSeconds *int
	Notes       *string
	Movements   []SaveMovementInput
}

// SaveMovementInput is one movement of a line; its position in the block is its order_in_block.
type SaveMovementInput struct {
	MovementID      string
	Reps            *int
	DurationSeconds *int
	Load            *string
}

// The int pointers below tell a missing value from 0, which is a valid order_index.

type AddDayInput struct {
	Label      string
	OrderIndex *int
}

type AddBlockInput struct {
	OrderIndex  *int
	Sets        *int
	RestSeconds *int
	Notes       *string
}

type AddMovementInput struct {
	MovementID      string
	Reps            *int
	DurationSeconds *int
	Load            *string
	OrderInBlock    *int
}

type Service struct {
	repo      *Repository
	athletes  *athlete.Service
	movements *movement.Service
}

func NewService(repo *Repository, athletes *athlete.Service, movements *movement.Service) *Service {
	return &Service{repo: repo, athletes: athletes, movements: movements}
}

// List returns athlete.ErrNotFound rather than an empty list for another coach's athlete.
func (s *Service) List(ctx context.Context, coachID, athleteID uuid.UUID) (*ListResult, error) {
	if _, err := s.athletes.Get(ctx, coachID, athleteID); err != nil {
		return nil, err
	}

	plans, err := s.repo.ListByAthlete(ctx, athleteID)
	if err != nil {
		return nil, err
	}

	sizes, err := s.repo.Sizes(ctx, planIDs(plans))
	if err != nil {
		return nil, err
	}

	return &ListResult{Plans: plans, CurrentID: currentPlanID(plans, time.Now()), Sizes: sizes}, nil
}

// CurrentResult holds one plan per athlete that has a current plan; the others are left out.
type CurrentResult struct {
	Plans []Plan
	Sizes map[uuid.UUID]Size
}

// Current finds the current plan of every one of the coach's athletes in a fixed number of queries.
// Ownership comes from the athlete list itself, so another coach's or a deleted athlete's plans never load.
func (s *Service) Current(ctx context.Context, coachID uuid.UUID) (*CurrentResult, error) {
	athleteIDs, err := s.athletes.IDs(ctx, coachID)
	if err != nil {
		return nil, err
	}

	plans, err := s.repo.ListByAthletes(ctx, athleteIDs)
	if err != nil {
		return nil, err
	}

	byAthlete := make(map[uuid.UUID][]Plan, len(athleteIDs))
	for _, p := range plans {
		byAthlete[p.AthleteID] = append(byAthlete[p.AthleteID], p)
	}

	now := time.Now()
	current := make([]Plan, 0, len(athleteIDs))
	for _, athleteID := range athleteIDs {
		athletePlans := byAthlete[athleteID]
		currentID := currentPlanID(athletePlans, now)
		for _, p := range athletePlans {
			if p.ID == currentID {
				current = append(current, p)
			}
		}
	}

	sizes, err := s.repo.Sizes(ctx, planIDs(current))
	if err != nil {
		return nil, err
	}

	return &CurrentResult{Plans: current, Sizes: sizes}, nil
}

func planIDs(plans []Plan) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(plans))
	for _, p := range plans {
		ids = append(ids, p.ID)
	}
	return ids
}

// Get returns ErrNotFound for another coach's plan, exactly as for one that doesn't exist.
func (s *Service) Get(ctx context.Context, coachID, planID uuid.UUID) (*Plan, error) {
	p, err := s.ownedPlan(ctx, coachID, planID)
	if err != nil {
		return nil, err
	}

	days, err := s.repo.LoadDays(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	p.Days = days

	if p.SourcePlanID != nil {
		source, err := s.repo.FindByID(ctx, *p.SourcePlanID)
		if err != nil && !errors.Is(err, ErrNotFound) {
			return nil, err
		}
		if source != nil {
			p.SourceTitle = &source.Title
		}
	}

	return p, nil
}

// Create validates before checking ownership: the athlete id is itself a body field and has to parse first.
// With a source plan it copies the source's days, blocks and block movements, and returns the full copy.
func (s *Service) Create(ctx context.Context, coachID uuid.UUID, input CreateInput) (*Plan, error) {
	p, err := buildPlan(input)
	if err != nil {
		return nil, err
	}

	if _, err := s.athletes.Get(ctx, coachID, p.AthleteID); err != nil {
		return nil, err
	}

	if p.SourcePlanID == nil {
		if err := s.repo.Create(ctx, p); err != nil {
			return nil, err
		}
		return p, nil
	}

	// The source may be any of the coach's athletes' plans; one the coach can't use is a field error
	// (docs/adr/0019), and the copy names its movements as the source does, deleted ones included.
	if _, err := s.ownedPlan(ctx, coachID, *p.SourcePlanID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, &ValidationError{Fields: map[string]string{fieldSourcePlanID: msgNotFound}}
		}
		return nil, err
	}

	days, err := s.repo.LoadDays(ctx, *p.SourcePlanID)
	if err != nil {
		return nil, err
	}

	if err := s.repo.CreateCopy(ctx, p, copyDays(p.ID, days)); err != nil {
		return nil, err
	}

	return s.Get(ctx, coachID, p.ID)
}

// copyDays gives the source's tree new ids, generated here so each child can point at its new parent
// before anything is inserted.
func copyDays(planID uuid.UUID, source []Day) []Day {
	days := make([]Day, 0, len(source))
	for _, d := range source {
		day := Day{ID: uuid.New(), PlanID: planID, Label: d.Label, OrderIndex: d.OrderIndex}
		for _, b := range d.Blocks {
			block := Block{
				ID: uuid.New(), DayID: day.ID, OrderIndex: b.OrderIndex,
				Sets: b.Sets, RestSeconds: b.RestSeconds, Notes: b.Notes,
			}
			for _, m := range b.Movements {
				block.Movements = append(block.Movements, BlockMovement{
					ID: uuid.New(), BlockID: block.ID, MovementID: m.MovementID,
					Reps: m.Reps, DurationSeconds: m.DurationSeconds, Load: m.Load, OrderInBlock: m.OrderInBlock,
				})
			}
			day.Blocks = append(day.Blocks, block)
		}
		days = append(days, day)
	}
	return days
}

// Update replaces title, note and start_date: an omitted note is cleared. A new start_date can change
// which plan is current; that is docs/adr/0016's rule doing its job.
func (s *Service) Update(ctx context.Context, coachID, planID uuid.UUID, input UpdateInput) (*Plan, error) {
	p, err := s.ownedPlan(ctx, coachID, planID)
	if err != nil {
		return nil, err
	}

	fields := make(map[string]string)
	p.StartDate, p.Title, p.Note = planFields(fields, input)
	if len(fields) > 0 {
		return nil, &ValidationError{Fields: fields}
	}

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}

	return s.Get(ctx, coachID, planID)
}

// Delete removes the plan with its days, blocks and block movements, through the FK cascades.
func (s *Service) Delete(ctx context.Context, coachID, planID uuid.UUID) error {
	if _, err := s.ownedPlan(ctx, coachID, planID); err != nil {
		return err
	}

	return s.repo.Delete(ctx, planID)
}

// AddDay checks ownership before validating, so another coach's plan is a 404 whatever the body holds.
func (s *Service) AddDay(ctx context.Context, coachID, planID uuid.UUID, input AddDayInput) (*Day, error) {
	if _, err := s.ownedPlan(ctx, coachID, planID); err != nil {
		return nil, err
	}

	d, err := buildDay(planID, input)
	if err != nil {
		return nil, err
	}

	if err := s.repo.CreateDay(ctx, d); err != nil {
		return nil, daySlotError(err)
	}

	return d, nil
}

// AddBlock checks ownership before validating, so another coach's day is a 404 whatever the body holds.
func (s *Service) AddBlock(ctx context.Context, coachID, dayID uuid.UUID, input AddBlockInput) (*Block, error) {
	athleteID, err := s.repo.AthleteIDOfDay(ctx, dayID)
	if err != nil {
		return nil, err
	}

	if err := s.authorize(ctx, coachID, athleteID); err != nil {
		return nil, err
	}

	b, err := buildBlock(dayID, input)
	if err != nil {
		return nil, err
	}

	if err := s.repo.CreateBlock(ctx, b); err != nil {
		return nil, err
	}

	return b, nil
}

// AddMovements is all or nothing: every entry must be valid and in the coach's library before any is inserted.
func (s *Service) AddMovements(ctx context.Context, coachID, blockID uuid.UUID, inputs []AddMovementInput) ([]BlockMovement, error) {
	athleteID, err := s.repo.AthleteIDOfBlock(ctx, blockID)
	if err != nil {
		return nil, err
	}

	if err := s.authorize(ctx, coachID, athleteID); err != nil {
		return nil, err
	}

	if len(inputs) == 0 {
		return nil, ErrNoMovements
	}

	fields := make(map[string]string)
	movements := make([]BlockMovement, 0, len(inputs))
	for i, input := range inputs {
		movements = append(movements, buildBlockMovement(fields, i, blockID, input))
	}

	if err := s.attachLibrary(ctx, coachID, movements, fields); err != nil {
		return nil, err
	}

	if len(fields) > 0 {
		return nil, &ValidationError{Fields: fields}
	}

	if err := s.repo.CreateBlockMovements(ctx, movements); err != nil {
		return nil, err
	}

	return movements, nil
}

// SaveDay replaces the day's blocks with the given ones, all or nothing, and returns the saved day.
// Block and block movement ids are new on every save. Errors are keyed by position, "[2].sets" or
// "[2].movements[1].movement_id", and reported together.
func (s *Service) SaveDay(ctx context.Context, coachID, dayID uuid.UUID, inputs []SaveBlockInput) (*Day, error) {
	athleteID, err := s.repo.AthleteIDOfDay(ctx, dayID)
	if err != nil {
		return nil, err
	}

	if err := s.authorize(ctx, coachID, athleteID); err != nil {
		return nil, err
	}

	if len(inputs) > maxBlocksPerDay {
		return nil, ErrTooManyBlocks
	}

	fields := make(map[string]string)
	blocks := make([]Block, 0, len(inputs))
	var movements []BlockMovement
	// movementKeys[k] is the field key of movements[k]'s movement_id.
	var movementKeys []string

	for i, input := range inputs {
		block := Block{
			ID:          uuid.New(),
			DayID:       dayID,
			OrderIndex:  i,
			Sets:        requiredInt(fields, elementField(i, "sets"), input.Sets, 1, maxSets),
			RestSeconds: input.RestSeconds,
			Notes:       optionalText(fields, elementField(i, "notes"), input.Notes, maxNotesLength),
		}
		optionalInt(fields, elementField(i, "rest_seconds"), input.RestSeconds, 0, maxRestSeconds)

		switch {
		case len(input.Movements) == 0:
			fields[elementField(i, "movements")] = msgNoMovements
		case len(input.Movements) > maxMovementsPerBlock:
			fields[elementField(i, "movements")] = fmt.Sprintf(msgTooMany, maxMovementsPerBlock)
		}

		for j, m := range input.Movements {
			prefix := fmt.Sprintf("[%d].movements[%d]", i, j)
			key := prefix + "." + fieldMovementID
			optionalInt(fields, prefix+".reps", m.Reps, 1, maxReps)
			optionalInt(fields, prefix+".duration_seconds", m.DurationSeconds, 1, maxDurationSeconds)
			movements = append(movements, BlockMovement{
				ID:              uuid.New(),
				BlockID:         block.ID,
				MovementID:      requiredUUID(fields, key, m.MovementID),
				Reps:            m.Reps,
				DurationSeconds: m.DurationSeconds,
				Load:            optionalText(fields, prefix+".load", m.Load, maxLoadLength),
				OrderInBlock:    j,
			})
			movementKeys = append(movementKeys, key)
		}

		blocks = append(blocks, block)
	}

	if err := s.checkDayMovements(ctx, coachID, dayID, movements, movementKeys, fields); err != nil {
		return nil, err
	}

	if len(fields) > 0 {
		return nil, &ValidationError{Fields: fields}
	}

	if err := s.repo.ReplaceBlocks(ctx, dayID, blocks, movements); err != nil {
		return nil, err
	}

	return s.repo.LoadDay(ctx, dayID)
}

// checkDayMovements flags every movement the coach can't use, without saying whether it is unknown or
// another coach's. A movement the day already uses passes even if it has left the library since, because
// the plan keeps naming a deleted movement (docs/adr/0022): otherwise deleting one would make every day
// that uses it impossible to save.
func (s *Service) checkDayMovements(ctx context.Context, coachID, dayID uuid.UUID, movements []BlockMovement, keys []string, fields map[string]string) error {
	ids := make([]uuid.UUID, 0, len(movements))
	for k := range movements {
		if _, flagged := fields[keys[k]]; !flagged {
			ids = append(ids, movements[k].MovementID)
		}
	}
	if len(ids) == 0 {
		return nil
	}

	usable := make(map[uuid.UUID]bool, len(ids))
	visible, err := s.movements.Visible(ctx, coachID, ids)
	if err != nil {
		return err
	}
	for _, m := range visible {
		usable[m.ID] = true
	}
	inDay, err := s.repo.MovementIDsOfDay(ctx, dayID)
	if err != nil {
		return err
	}
	for _, id := range inDay {
		usable[id] = true
	}

	for k := range movements {
		if _, flagged := fields[keys[k]]; !flagged && !usable[movements[k].MovementID] {
			fields[keys[k]] = msgNotFound
		}
	}
	return nil
}

// DeleteDay removes a day and its blocks; the plan's other days keep their labels and slots, so the
// home screen's letters keep their columns.
func (s *Service) DeleteDay(ctx context.Context, coachID, dayID uuid.UUID) error {
	athleteID, err := s.repo.AthleteIDOfDay(ctx, dayID)
	if err != nil {
		return err
	}

	if err := s.authorize(ctx, coachID, athleteID); err != nil {
		return err
	}

	return s.repo.DeleteDay(ctx, dayID)
}

func (s *Service) ownedPlan(ctx context.Context, coachID, planID uuid.UUID) (*Plan, error) {
	p, err := s.repo.FindByID(ctx, planID)
	if err != nil {
		return nil, err
	}

	if err := s.authorize(ctx, coachID, p.AthleteID); err != nil {
		return nil, err
	}

	return p, nil
}

// authorize is where a plan, day or block meets tenant scoping, through the athlete at the top of its chain.
func (s *Service) authorize(ctx context.Context, coachID, athleteID uuid.UUID) error {
	if _, err := s.athletes.Get(ctx, coachID, athleteID); err != nil {
		if errors.Is(err, athlete.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

// attachLibrary sets each entry's library movement and flags the ones the coach can't use, without
// saying whether the id is unknown or another coach's. Entries whose id didn't parse are already flagged.
func (s *Service) attachLibrary(ctx context.Context, coachID uuid.UUID, movements []BlockMovement, fields map[string]string) error {
	ids := make([]uuid.UUID, 0, len(movements))
	for i := range movements {
		if _, flagged := fields[elementField(i, fieldMovementID)]; !flagged {
			ids = append(ids, movements[i].MovementID)
		}
	}

	if len(ids) == 0 {
		return nil
	}

	found, err := s.movements.Visible(ctx, coachID, ids)
	if err != nil {
		return err
	}

	library := make(map[uuid.UUID]movement.Movement, len(found))
	for _, m := range found {
		library[m.ID] = m
	}

	for i := range movements {
		key := elementField(i, fieldMovementID)
		if _, flagged := fields[key]; flagged {
			continue
		}

		m, ok := library[movements[i].MovementID]
		if !ok {
			fields[key] = msgNotFound
			continue
		}
		movements[i].Movement = m
	}

	return nil
}

// daySlotError reports a slot the schema rejected against order_index, in the same words buildDay uses.
func daySlotError(err error) error {
	switch {
	case errors.Is(err, ErrDaySlotOutOfRange):
		return &ValidationError{Fields: map[string]string{fieldOrderIndex: fmt.Sprintf(msgRange, 0, MaxDaysPerPlan-1)}}
	case errors.Is(err, ErrDaySlotTaken):
		return &ValidationError{Fields: map[string]string{fieldOrderIndex: msgDaySlotTaken}}
	default:
		return err
	}
}

// currentPlanID picks the plan the athlete is on today: the latest one whose start_date is not
// in the future. A future-dated plan is upcoming, not current, so writing next block's plan early
// doesn't hide the one being trained. "Today" is the server's local date. plans must be in
// ListByAthlete's order, so the first match wins and same-day ties go to the later-created plan.
func currentPlanID(plans []Plan, now time.Time) uuid.UUID {
	// DATE columns scan as midnight UTC, so today is built the same way to compare dates, not instants.
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	for _, p := range plans {
		if !p.StartDate.After(today) {
			return p.ID
		}
	}

	return uuid.Nil
}

// buildPlan collects every field error before returning, so a coach sees all mistakes at once.
func buildPlan(input CreateInput) (*Plan, error) {
	fields := make(map[string]string)

	athleteID := requiredUUID(fields, "athlete_id", input.AthleteID)
	startDate, title, note := planFields(fields, UpdateInput{StartDate: input.StartDate, Title: input.Title, Note: input.Note})

	var sourceID *uuid.UUID
	if strings.TrimSpace(input.SourcePlanID) != "" {
		id := requiredUUID(fields, fieldSourcePlanID, input.SourcePlanID)
		sourceID = &id
	}

	if len(fields) > 0 {
		return nil, &ValidationError{Fields: fields}
	}

	return &Plan{
		ID:           uuid.New(),
		AthleteID:    athleteID,
		SourcePlanID: sourceID,
		StartDate:    startDate,
		Title:        title,
		Note:         note,
	}, nil
}

// planFields checks the fields a plan's create and update share.
func planFields(fields map[string]string, input UpdateInput) (time.Time, string, *string) {
	rawDate := strings.TrimSpace(input.StartDate)
	startDate, err := time.Parse(time.DateOnly, rawDate)
	switch {
	case rawDate == "":
		fields["start_date"] = httpx.MsgFieldRequired
	case err != nil:
		fields["start_date"] = httpx.MsgDateFormat
	}

	title := requiredText(fields, "title", input.Title, maxTitleLength)
	note := optionalText(fields, "note", input.Note, maxNotesLength)

	return startDate, title, note
}

func buildDay(planID uuid.UUID, input AddDayInput) (*Day, error) {
	fields := make(map[string]string)

	label := strings.TrimSpace(input.Label)
	switch {
	case label == "":
		fields["label"] = httpx.MsgFieldRequired
	case !slices.Contains(dayLabels, label):
		fields["label"] = msgDayLabel
	}

	// The schema enforces the slot range as well; checking here too reports it alongside the label.
	orderIndex := requiredInt(fields, fieldOrderIndex, input.OrderIndex, 0, MaxDaysPerPlan-1)

	if len(fields) > 0 {
		return nil, &ValidationError{Fields: fields}
	}

	return &Day{ID: uuid.New(), PlanID: planID, Label: label, OrderIndex: orderIndex}, nil
}

func buildBlock(dayID uuid.UUID, input AddBlockInput) (*Block, error) {
	fields := make(map[string]string)

	orderIndex := requiredInt(fields, fieldOrderIndex, input.OrderIndex, 0, maxPosition)
	sets := requiredInt(fields, "sets", input.Sets, 1, maxSets)
	optionalInt(fields, "rest_seconds", input.RestSeconds, 0, maxRestSeconds)
	notes := optionalText(fields, "notes", input.Notes, maxNotesLength)

	if len(fields) > 0 {
		return nil, &ValidationError{Fields: fields}
	}

	return &Block{
		ID:          uuid.New(),
		DayID:       dayID,
		OrderIndex:  orderIndex,
		Sets:        sets,
		RestSeconds: input.RestSeconds,
		Notes:       notes,
	}, nil
}

// buildBlockMovement records its errors under "[i].name" so one map covers the whole batch.
func buildBlockMovement(fields map[string]string, i int, blockID uuid.UUID, input AddMovementInput) BlockMovement {
	movementID := requiredUUID(fields, elementField(i, fieldMovementID), input.MovementID)
	optionalInt(fields, elementField(i, "reps"), input.Reps, 1, maxReps)
	optionalInt(fields, elementField(i, "duration_seconds"), input.DurationSeconds, 1, maxDurationSeconds)
	load := optionalText(fields, elementField(i, "load"), input.Load, maxLoadLength)
	orderInBlock := requiredInt(fields, elementField(i, "order_in_block"), input.OrderInBlock, 0, maxPosition)

	return BlockMovement{
		ID:              uuid.New(),
		BlockID:         blockID,
		MovementID:      movementID,
		Reps:            input.Reps,
		DurationSeconds: input.DurationSeconds,
		Load:            load,
		OrderInBlock:    orderInBlock,
	}
}

func elementField(i int, name string) string {
	return fmt.Sprintf("[%d].%s", i, name)
}

func requiredUUID(fields map[string]string, name, value string) uuid.UUID {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		fields[name] = httpx.MsgFieldRequired
		return uuid.Nil
	}

	id, err := uuid.Parse(trimmed)
	if err != nil {
		fields[name] = msgUUID
	}
	return id
}

func requiredText(fields map[string]string, name, value string, maxLength int) string {
	trimmed := strings.TrimSpace(value)
	switch {
	case trimmed == "":
		fields[name] = httpx.MsgFieldRequired
	case utf8.RuneCountInString(trimmed) > maxLength:
		fields[name] = fmt.Sprintf(msgMaxLength, maxLength)
	}
	return trimmed
}

// optionalText collapses blank input to nil so it is stored as NULL, not "".
func optionalText(fields map[string]string, name string, value *string, maxLength int) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	if utf8.RuneCountInString(trimmed) > maxLength {
		fields[name] = fmt.Sprintf(msgMaxLength, maxLength)
	}
	return &trimmed
}

// requiredInt returns 0 for a missing value; the error it records keeps that 0 from being stored.
func requiredInt(fields map[string]string, name string, value *int, lo, hi int) int {
	if value == nil {
		fields[name] = httpx.MsgFieldRequired
		return 0
	}

	optionalInt(fields, name, value, lo, hi)
	return *value
}

func optionalInt(fields map[string]string, name string, value *int, lo, hi int) {
	if value != nil && (*value < lo || *value > hi) {
		fields[name] = fmt.Sprintf(msgRange, lo, hi)
	}
}
