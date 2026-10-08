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
	"github.com/AbolfazlTayyar/gogym/internal/httpx"
	"github.com/AbolfazlTayyar/gogym/internal/movement"
	"github.com/google/uuid"
)

// Length limits count runes, not bytes, so Persian text gets the same allowance as Latin.
const (
	maxTitleLength = 100
	maxNotesLength = 2000
	maxLoadLength  = 100
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
	fieldOrderIndex = "order_index"
	fieldMovementID = "movement_id"
)

const (
	msgUUID             = "must be a UUID"
	msgMaxLength        = "must be at most %d characters"
	msgRange            = "must be between %d and %d"
	msgDayLabel         = "must be one of A-G or day1-day7"
	msgDaySlotTaken     = "is already used by another day in this plan"
	msgMovementNotFound = "not found"
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

// ListResult's CurrentID is uuid.Nil when the athlete has no plan that has started yet.
type ListResult struct {
	Plans     []Plan
	CurrentID uuid.UUID
}

// CreateInput's AthleteID is a string so a malformed id is reported as a field error like any other.
type CreateInput struct {
	AthleteID string
	StartDate string
	Title     string
	Note      *string
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
	repo     *Repository
	athletes *athlete.Service
}

func NewService(repo *Repository, athletes *athlete.Service) *Service {
	return &Service{repo: repo, athletes: athletes}
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

	return &ListResult{Plans: plans, CurrentID: currentPlanID(plans, time.Now())}, nil
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

	return p, nil
}

// Create validates before checking ownership: the athlete id is itself a body field and has to parse first.
func (s *Service) Create(ctx context.Context, coachID uuid.UUID, input CreateInput) (*Plan, error) {
	p, err := buildPlan(input)
	if err != nil {
		return nil, err
	}

	if _, err := s.athletes.Get(ctx, coachID, p.AthleteID); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
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

	found, err := s.repo.FindVisibleMovements(ctx, coachID, ids)
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
			fields[key] = msgMovementNotFound
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

	if len(fields) > 0 {
		return nil, &ValidationError{Fields: fields}
	}

	return &Plan{
		ID:        uuid.New(),
		AthleteID: athleteID,
		StartDate: startDate,
		Title:     title,
		Note:      note,
	}, nil
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
