package movement

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"
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

// ErrUniversal is for a movement the coach can see but not change; one they can't see is ErrNotFound.
var ErrUniversal = errors.New("movement: universal movements are read-only")

// ErrUniversalCategory is for a category a universal movement is filed under: renaming or clearing it
// would change only the coach's own movements and split the category in two.
var ErrUniversalCategory = errors.New("movement: universal categories are read-only")

// Arabic keyboards type ي, ى and ك where Persian has ی and ک, so the same word typed on each would
// otherwise be two categories that look alike.
var arabicLetters = strings.NewReplacer("ي", "ی", "ى", "ی", "ك", "ک")

const zwnj = '‌'

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

// RenameCategory refiles every one of the coach's movements under from as newName and returns the name
// as stored. Renaming onto a category already in use merges the two, which is how a duplicate goes.
func (s *Service) RenameCategory(ctx context.Context, coachID uuid.UUID, from, newName string) (string, error) {
	from = normalizeCategory(from)
	if err := s.ownedCategory(ctx, coachID, from); err != nil {
		return "", err
	}

	fields := make(map[string]string)
	to := requiredText(fields, "name", normalizeCategory(newName), maxCategoryLength)
	if len(fields) > 0 {
		return "", &ValidationError{Fields: fields}
	}

	if err := s.repo.SetCategory(ctx, coachID, from, &to); err != nil {
		return "", err
	}

	return to, nil
}

// DeleteCategory leaves every one of the coach's movements filed under name without a category.
func (s *Service) DeleteCategory(ctx context.Context, coachID uuid.UUID, name string) error {
	name = normalizeCategory(name)
	if err := s.ownedCategory(ctx, coachID, name); err != nil {
		return err
	}

	return s.repo.SetCategory(ctx, coachID, name, nil)
}

// ownedCategory only rules out a universal category; SetCategory reports ErrNotFound when none of the
// coach's movements uses the name.
func (s *Service) ownedCategory(ctx context.Context, coachID uuid.UUID, name string) error {
	universal, err := s.repo.CategoryIsUniversal(ctx, coachID, name)
	if err != nil {
		return err
	}

	if universal {
		return ErrUniversalCategory
	}

	return nil
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
	category := boundedText(fields, "category", optionalCategory(input.Category), maxCategoryLength)
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
		fields[name] = fmt.Sprintf(httpx.MsgMaxLength, maxLength)
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

// normalizeCategory makes spellings that look alike one category: Arabic letters become Persian, runs
// of spaces collapse to one, and spaces and ZWNJs at either end go.
func normalizeCategory(value string) string {
	collapsed := strings.Join(strings.Fields(arabicLetters.Replace(value)), " ")
	return strings.TrimFunc(collapsed, func(r rune) bool { return unicode.IsSpace(r) || r == zwnj })
}

func optionalCategory(value *string) *string {
	if value == nil {
		return nil
	}

	normalized := normalizeCategory(*value)
	return optionalText(&normalized)
}

func boundedText(fields map[string]string, name string, value *string, maxLength int) *string {
	if value != nil && utf8.RuneCountInString(*value) > maxLength {
		fields[name] = fmt.Sprintf(httpx.MsgMaxLength, maxLength)
	}
	return value
}
