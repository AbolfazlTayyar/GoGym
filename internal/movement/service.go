package movement

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/AbolfazlTayyar/gogym/internal/platform/httpx"
	"github.com/google/uuid"
)

// Length limits count runes, not bytes, so Persian text gets the same allowance as Latin.
const (
	maxNameLength        = 100
	maxCategoryLength    = 50
	maxDescriptionLength = 2000
)

const msgMaxLength = "must be at most %d characters"

// ErrUniversal is for a movement the coach can see but not change; one they can't see is ErrNotFound.
var ErrUniversal = errors.New("movement: universal movements are read-only")

// ValidationError's Fields are keyed by the json names the client sent.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("movement: validation failed for %d field(s)", len(e.Fields))
}

// Input is a new movement or an update's full replacement. It has no CoachID, so a request body can
// never choose the owning coach.
type Input struct {
	Name        string
	Category    *string
	Description *string
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// List returns the coach's own movements and the universal ones together, by name; a blank nameQuery
// matches every movement.
func (s *Service) List(ctx context.Context, coachID uuid.UUID, nameQuery string) ([]Movement, error) {
	return s.repo.List(ctx, coachID, nameQuery)
}

// Visible is how other modules check movement ids against the coach's library: it returns the ones
// the coach can use and silently drops the rest, so a caller can't tell another coach's from unknown.
func (s *Service) Visible(ctx context.Context, coachID uuid.UUID, ids []uuid.UUID) ([]Movement, error) {
	return s.repo.FindVisibleByIDs(ctx, coachID, ids)
}

func (s *Service) Create(ctx context.Context, coachID uuid.UUID, input Input) (*Movement, error) {
	m := &Movement{ID: uuid.New(), CoachID: &coachID}
	if err := applyInput(m, input); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, m); err != nil {
		return nil, err
	}

	return m, nil
}

// Update checks the movement before validating, so another coach's is a 404 and a universal one a 403
// whatever the body holds.
func (s *Service) Update(ctx context.Context, coachID, id uuid.UUID, input Input) (*Movement, error) {
	m, err := s.owned(ctx, coachID, id)
	if err != nil {
		return nil, err
	}

	if err := applyInput(m, input); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, coachID, m); err != nil {
		return nil, err
	}

	return m, nil
}

// Delete goes ahead even when a plan uses the movement, rather than refusing with a 409: the delete is
// soft and plan detail loads movements unscoped, so existing plans keep naming it while the library and
// new picks lose it. Refusing would mean a movement used in one old plan could never leave the library.
func (s *Service) Delete(ctx context.Context, coachID, id uuid.UUID) error {
	if _, err := s.owned(ctx, coachID, id); err != nil {
		return err
	}

	return s.repo.Delete(ctx, coachID, id)
}

// owned answers a universal movement with ErrUniversal: the coach already sees it in their list, so
// saying it exists leaks nothing, unlike for another coach's movement.
func (s *Service) owned(ctx context.Context, coachID, id uuid.UUID) (*Movement, error) {
	m, err := s.repo.FindVisible(ctx, coachID, id)
	if err != nil {
		return nil, err
	}

	if m.IsUniversal() {
		return nil, ErrUniversal
	}

	return m, nil
}

// applyInput collects every field error before returning, so a coach sees all mistakes at once, and
// leaves m untouched unless the whole input is valid.
func applyInput(m *Movement, input Input) error {
	fields := make(map[string]string)

	name := requiredText(fields, "name", input.Name, maxNameLength)
	category := boundedText(fields, "category", optionalText(input.Category), maxCategoryLength)
	description := boundedText(fields, "description", optionalText(input.Description), maxDescriptionLength)

	if len(fields) > 0 {
		return &ValidationError{Fields: fields}
	}

	m.Name = name
	m.Category = category
	m.Description = description

	return nil
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

func boundedText(fields map[string]string, name string, value *string, maxLength int) *string {
	if value != nil && utf8.RuneCountInString(*value) > maxLength {
		fields[name] = fmt.Sprintf(msgMaxLength, maxLength)
	}
	return value
}
