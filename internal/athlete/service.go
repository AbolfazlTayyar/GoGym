package athlete

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/AbolfazlTayyar/gogym/internal/httpx"
	"github.com/AbolfazlTayyar/gogym/internal/validate"
	"github.com/google/uuid"
)

// Length limits count runes, not bytes, so Persian text gets the same allowance as Latin.
const (
	maxNameLength  = 100
	maxNotesLength = 2000
)

// Height bounds in centimetres, tight enough to catch a client that sent metres.
const (
	minHeightCM = 50
	maxHeightCM = 250
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

// TypeFilterAll is a filter value only, never stored in athlete_type.
const TypeFilterAll = "all"

// ValidationError's Fields are keyed by the json names the client sent.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("athlete: validation failed for %d field(s)", len(e.Fields))
}

// CreateInput deliberately has no CoachID, so a request body can never choose the owning coach.
type CreateInput struct {
	FirstName       string
	LastName        string
	Phone           string
	ExperienceLevel *string
	Injuries        *string
	Goal            *string
	Height          *float64
	// AthleteType empty means TypePrivate.
	AthleteType string
}

type ListOptions struct {
	Query string
	// Type empty means TypePrivate, the coach's working list.
	Type   string
	Limit  int
	Offset int
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Create validates here rather than via binding tags so the rules hold for every caller, not just HTTP.
func (s *Service) Create(ctx context.Context, coachID uuid.UUID, input CreateInput) (*Athlete, error) {
	a, err := buildAthlete(coachID, input)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, a); err != nil {
		return nil, err
	}

	return a, nil
}

// Get returns ErrNotFound for another coach's athlete, exactly as for one that doesn't exist.
func (s *Service) Get(ctx context.Context, coachID, id uuid.UUID) (*Athlete, error) {
	return s.repo.FindByID(ctx, coachID, id)
}

// ListResult's Limit and Offset are the window actually applied, after defaults and clamping.
type ListResult struct {
	Athletes []Athlete
	Total    int64
	Limit    int
	Offset   int
}

func (s *Service) List(ctx context.Context, coachID uuid.UUID, opts ListOptions) (*ListResult, error) {
	normalized, err := normalizeListOptions(opts)
	if err != nil {
		return nil, err
	}

	athletes, total, err := s.repo.List(ctx, coachID, normalized)
	if err != nil {
		return nil, err
	}

	return &ListResult{
		Athletes: athletes,
		Total:    total,
		Limit:    normalized.Limit,
		Offset:   normalized.Offset,
	}, nil
}

// buildAthlete collects every field error before returning, so a coach sees all mistakes at once.
func buildAthlete(coachID uuid.UUID, input CreateInput) (*Athlete, error) {
	fields := make(map[string]string)

	firstName := requiredName(fields, "first_name", input.FirstName)
	lastName := requiredName(fields, "last_name", input.LastName)

	phone := strings.TrimSpace(input.Phone)
	switch {
	case phone == "":
		fields["phone"] = httpx.MsgFieldRequired
	case !validate.IsIranMobile(phone):
		fields["phone"] = validate.MsgIranMobile
	}

	experience := optionalText(input.ExperienceLevel)
	if experience != nil {
		level := strings.ToLower(*experience)
		switch level {
		case ExperienceBeginner, ExperienceIntermediate, ExperienceAdvanced:
			experience = &level
		default:
			fields["experience_level"] = fmt.Sprintf("must be one of %s, %s, %s",
				ExperienceBeginner, ExperienceIntermediate, ExperienceAdvanced)
		}
	}

	injuries := boundedText(fields, "injuries", optionalText(input.Injuries))
	goal := boundedText(fields, "goal", optionalText(input.Goal))

	if input.Height != nil && (*input.Height < minHeightCM || *input.Height > maxHeightCM) {
		fields["height"] = fmt.Sprintf("must be between %d and %d centimetres", minHeightCM, maxHeightCM)
	}

	athleteType := strings.TrimSpace(input.AthleteType)
	switch athleteType {
	case "":
		athleteType = TypePrivate
	case TypePrivate, TypePublic:
	default:
		fields["athlete_type"] = fmt.Sprintf("must be %s or %s", TypePrivate, TypePublic)
	}

	if len(fields) > 0 {
		return nil, &ValidationError{Fields: fields}
	}

	return &Athlete{
		ID:              uuid.New(),
		CoachID:         coachID,
		FirstName:       firstName,
		LastName:        lastName,
		Phone:           phone,
		ExperienceLevel: experience,
		Injuries:        injuries,
		Goal:            goal,
		Height:          input.Height,
		AthleteType:     athleteType,
	}, nil
}

func requiredName(fields map[string]string, name, value string) string {
	trimmed := strings.TrimSpace(value)
	switch {
	case trimmed == "":
		fields[name] = httpx.MsgFieldRequired
	case utf8.RuneCountInString(trimmed) > maxNameLength:
		fields[name] = fmt.Sprintf("must be at most %d characters", maxNameLength)
	}
	return trimmed
}

// optionalText collapses blank input to nil so it is stored as NULL, not "".
func optionalText(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

func boundedText(fields map[string]string, name string, value *string) *string {
	if value != nil && utf8.RuneCountInString(*value) > maxNotesLength {
		fields[name] = fmt.Sprintf("must be at most %d characters", maxNotesLength)
	}
	return value
}

// normalizeListOptions rejects an unknown type rather than falling back, so a typo isn't silently served private.
func normalizeListOptions(opts ListOptions) (ListOptions, error) {
	out := opts

	out.Query = strings.TrimSpace(opts.Query)

	switch strings.ToLower(strings.TrimSpace(opts.Type)) {
	case "":
		out.Type = TypePrivate
	case TypePrivate:
		out.Type = TypePrivate
	case TypePublic:
		out.Type = TypePublic
	case TypeFilterAll:
		out.Type = TypeFilterAll
	default:
		return ListOptions{}, &ValidationError{Fields: map[string]string{
			"athlete_type": fmt.Sprintf("must be %s, %s, or %s", TypePrivate, TypePublic, TypeFilterAll),
		}}
	}

	if out.Limit <= 0 {
		out.Limit = defaultLimit
	}
	if out.Limit > maxLimit {
		out.Limit = maxLimit
	}
	if out.Offset < 0 {
		out.Offset = 0
	}

	return out, nil
}
