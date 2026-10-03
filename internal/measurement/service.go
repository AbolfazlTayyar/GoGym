package measurement

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/AbolfazlTayyar/gogym/internal/athlete"
	"github.com/AbolfazlTayyar/gogym/internal/httpx"
	"github.com/google/uuid"
)

// dateLayout is the wire format both ways, so a chart can use the value without parsing a timestamp.
const dateLayout = "2006-01-02"

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

const (
	msgDateFormat = "must be a date in YYYY-MM-DD format"
	msgNoValues   = "at least one of weight, chest, waist, arm, thigh or hip is required"
)

// ValidationError's Fields are keyed by the json names the client sent.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("measurement: validation failed for %d field(s)", len(e.Fields))
}

// CreateInput has no AthleteID: it comes from the path, after the service has checked who owns it.
type CreateInput struct {
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
func (s *Service) Create(ctx context.Context, coachID, athleteID uuid.UUID, input CreateInput) (*Measurement, error) {
	if _, err := s.athletes.Get(ctx, coachID, athleteID); err != nil {
		return nil, err
	}

	m, err := buildMeasurement(athleteID, input)
	if err != nil {
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

type metric struct {
	name     string
	value    *float64
	min, max float64
	unit     string
}

// buildMeasurement collects every field error before returning, so a coach sees all mistakes at once.
func buildMeasurement(athleteID uuid.UUID, input CreateInput) (*Measurement, error) {
	fields := make(map[string]string)

	rawDate := strings.TrimSpace(input.Date)
	date, err := time.Parse(dateLayout, rawDate)
	switch {
	case rawDate == "":
		fields["date"] = httpx.MsgFieldRequired
	case err != nil:
		fields["date"] = msgDateFormat
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
	for _, m := range metrics {
		if m.value == nil {
			continue
		}
		anyValue = true
		if *m.value < m.min || *m.value > m.max {
			fields[m.name] = fmt.Sprintf("must be between %g and %g %s", m.min, m.max, m.unit)
		}
	}

	// An all-empty check-in would plot as a gap, so every metric field is flagged for the form to highlight.
	if !anyValue {
		for _, m := range metrics {
			fields[m.name] = msgNoValues
		}
	}

	if len(fields) > 0 {
		return nil, &ValidationError{Fields: fields}
	}

	return &Measurement{
		ID:        uuid.New(),
		AthleteID: athleteID,
		Date:      date,
		Weight:    input.Weight,
		Chest:     input.Chest,
		Waist:     input.Waist,
		Arm:       input.Arm,
		Thigh:     input.Thigh,
		Hip:       input.Hip,
	}, nil
}
