package measurement

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/AbolfazlTayyar/gogym/internal/athlete"
	"github.com/AbolfazlTayyar/gogym/internal/platform/httpx"
	"github.com/google/uuid"
)

// dateLayout is the wire format both ways, so a chart can use the value without parsing a timestamp.
const dateLayout = "2006-01-02"

// The date window catches a mistyped year. It reaches a day past today in UTC because a coach east of
// UTC is already on tomorrow's date for part of every day.
var earliestDate = time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)

const maxDaysAhead = 1

const msgDateRange = "must be between %s and %s"

// Bounds are tight enough to catch a client that sent grams or metres, not to judge the athlete.
const (
	minWeightKG        = 20
	maxWeightKG        = 400
	minCircumferenceCM = 10
	maxCircumferenceCM = 250
)

const (
	unitKilograms   = "kilograms"
	unitCentimetres = "centimetres"
)

const msgNoValues = "at least one of weight, chest, waist, arm, thigh or hip is required"

// ValidationError's Fields are keyed by the json names the client sent.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("measurement: validation failed for %d field(s)", len(e.Fields))
}

// Input is a new measurement or an update's full replacement. It has no AthleteID: create takes it from
// the path once the athlete is checked, and an update can't move a measurement to another athlete.
type Input struct {
	Date   string
	Weight *float64
	Chest  *float64
	Waist  *float64
	Arm    *float64
	Thigh  *float64
	Hip    *float64
}

type Service struct {
	repo     *Repository
	athletes *athlete.Service
}

func NewService(repo *Repository, athletes *athlete.Service) *Service {
	return &Service{repo: repo, athletes: athletes}
}

// Create checks ownership before validating, so another coach's athlete is a 404 whatever the body holds.
func (s *Service) Create(ctx context.Context, coachID, athleteID uuid.UUID, input Input) (*Measurement, error) {
	if _, err := s.athletes.Get(ctx, coachID, athleteID); err != nil {
		return nil, err
	}

	m := &Measurement{ID: uuid.New(), AthleteID: athleteID}
	if err := applyInput(m, input, time.Now()); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, m); err != nil {
		return nil, err
	}

	return m, nil
}

// List returns athlete.ErrNotFound rather than an empty list for another coach's athlete.
func (s *Service) List(ctx context.Context, coachID, athleteID uuid.UUID) ([]Measurement, error) {
	if _, err := s.athletes.Get(ctx, coachID, athleteID); err != nil {
		return nil, err
	}

	return s.repo.ListByAthlete(ctx, athleteID)
}

// Update checks the measurement's athlete before validating, so another coach's measurement is a 404
// whatever the body holds.
func (s *Service) Update(ctx context.Context, coachID, id uuid.UUID, input Input) (*Measurement, error) {
	m, err := s.owned(ctx, coachID, id)
	if err != nil {
		return nil, err
	}

	if err := applyInput(m, input, time.Now()); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, m); err != nil {
		return nil, err
	}

	return m, nil
}

// Delete is permanent, unlike an athlete's or a movement's: nothing references a measurement, so a kept
// row would only serve an undo that doesn't exist.
func (s *Service) Delete(ctx context.Context, coachID, id uuid.UUID) error {
	m, err := s.owned(ctx, coachID, id)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, m)
}

// owned reaches tenant scoping through the measurement's athlete, so another coach's measurement, and one
// whose athlete was deleted, come back as athlete.ErrNotFound.
func (s *Service) owned(ctx context.Context, coachID, id uuid.UUID) (*Measurement, error) {
	m, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if _, err := s.athletes.Get(ctx, coachID, m.AthleteID); err != nil {
		return nil, err
	}

	return m, nil
}

type metric struct {
	name     string
	value    *float64
	min, max float64
	unit     string
}

// applyInput collects every field error before returning, so a coach sees all mistakes at once, and
// leaves m untouched unless the whole input is valid. now is a parameter so tests can pin the date window.
func applyInput(m *Measurement, input Input, now time.Time) error {
	fields := make(map[string]string)

	y, mo, d := now.UTC().Date()
	latestDate := time.Date(y, mo, d+maxDaysAhead, 0, 0, 0, 0, time.UTC)

	rawDate := strings.TrimSpace(input.Date)
	date, err := time.Parse(dateLayout, rawDate)
	switch {
	case rawDate == "":
		fields["date"] = httpx.MsgFieldRequired
	case err != nil:
		fields["date"] = httpx.MsgDateFormat
	case date.Before(earliestDate) || date.After(latestDate):
		fields["date"] = fmt.Sprintf(msgDateRange, earliestDate.Format(dateLayout), latestDate.Format(dateLayout))
	}

	metrics := []metric{
		{name: "weight", value: input.Weight, min: minWeightKG, max: maxWeightKG, unit: unitKilograms},
		{name: "chest", value: input.Chest, min: minCircumferenceCM, max: maxCircumferenceCM, unit: unitCentimetres},
		{name: "waist", value: input.Waist, min: minCircumferenceCM, max: maxCircumferenceCM, unit: unitCentimetres},
		{name: "arm", value: input.Arm, min: minCircumferenceCM, max: maxCircumferenceCM, unit: unitCentimetres},
		{name: "thigh", value: input.Thigh, min: minCircumferenceCM, max: maxCircumferenceCM, unit: unitCentimetres},
		{name: "hip", value: input.Hip, min: minCircumferenceCM, max: maxCircumferenceCM, unit: unitCentimetres},
	}

	anyValue := false
	for _, mt := range metrics {
		if mt.value == nil {
			continue
		}
		anyValue = true
		if *mt.value < mt.min || *mt.value > mt.max {
			fields[mt.name] = fmt.Sprintf("must be between %g and %g %s", mt.min, mt.max, mt.unit)
		}
	}

	// An all-empty check-in would plot as a gap, so every metric field is flagged for the form to highlight.
	if !anyValue {
		for _, mt := range metrics {
			fields[mt.name] = msgNoValues
		}
	}

	if len(fields) > 0 {
		return &ValidationError{Fields: fields}
	}

	m.Date = date
	m.Weight = input.Weight
	m.Chest = input.Chest
	m.Waist = input.Waist
	m.Arm = input.Arm
	m.Thigh = input.Thigh
	m.Hip = input.Hip

	return nil
}
