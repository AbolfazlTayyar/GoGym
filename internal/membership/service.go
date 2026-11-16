package membership

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AbolfazlTayyar/gogym/internal/plan"
	"github.com/AbolfazlTayyar/gogym/internal/platform/httpx"
	"github.com/google/uuid"
)

const dateLayout = "2006-01-02"

const fieldDate = "date"

// A visit's date is the phone's today, which runs up to a day ahead of the server's UTC date east of UTC.
const maxDaysAhead = 1

const (
	msgDateRange     = "must be between %s and %s"
	msgDuplicateDate = "is already recorded"
	msgSessionsUsed  = "all %d sessions are recorded"
)

// ValidationError's Fields are keyed by the json names the client sent.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("membership: validation failed for %d field(s)", len(e.Fields))
}

type Service struct {
	repo  *Repository
	plans *plan.Service
}

func NewService(repo *Repository, plans *plan.Service) *Service {
	return &Service{repo: repo, plans: plans}
}

// Get answers an untracked plan with Tracked false, not an error, so the plan page can show the toggle off.
func (s *Service) Get(ctx context.Context, coachID, planID uuid.UUID) (*Attendance, error) {
	if _, err := s.plans.Get(ctx, coachID, planID); err != nil {
		return nil, err
	}
	return s.repo.Find(ctx, planID)
}

func (s *Service) Track(ctx context.Context, coachID, planID uuid.UUID) (*Attendance, error) {
	if _, err := s.plans.Get(ctx, coachID, planID); err != nil {
		return nil, err
	}
	if err := s.repo.Track(ctx, planID); err != nil {
		return nil, err
	}
	return s.repo.Find(ctx, planID)
}

// Untrack deletes the plan's visits too; the client confirms that with the coach first.
func (s *Service) Untrack(ctx context.Context, coachID, planID uuid.UUID) (*Attendance, error) {
	if _, err := s.plans.Get(ctx, coachID, planID); err != nil {
		return nil, err
	}
	if err := s.repo.Untrack(ctx, planID); err != nil {
		return nil, err
	}
	return s.repo.Find(ctx, planID)
}

func (s *Service) AddVisit(ctx context.Context, coachID, planID uuid.UUID, rawDate string) (*Attendance, error) {
	p, err := s.plans.Get(ctx, coachID, planID)
	if err != nil {
		return nil, err
	}

	date, err := visitDate(rawDate, p.StartDate, time.Now())
	if err != nil {
		return nil, err
	}

	switch err := s.repo.AddVisit(ctx, planID, date); {
	case errors.Is(err, ErrDuplicateDate):
		return nil, &ValidationError{Fields: map[string]string{fieldDate: msgDuplicateDate}}
	case errors.Is(err, ErrSessionsUsed):
		return nil, &ValidationError{Fields: map[string]string{fieldDate: fmt.Sprintf(msgSessionsUsed, SessionsPerPlan)}}
	case err != nil:
		return nil, err
	}

	return s.repo.Find(ctx, planID)
}

// DeleteVisit answers a malformed date as a missing visit: it comes from the path, not a form field.
func (s *Service) DeleteVisit(ctx context.Context, coachID, planID uuid.UUID, rawDate string) (*Attendance, error) {
	if _, err := s.plans.Get(ctx, coachID, planID); err != nil {
		return nil, err
	}

	date, err := time.Parse(dateLayout, rawDate)
	if err != nil {
		return nil, ErrVisitNotFound
	}

	if err := s.repo.DeleteVisit(ctx, planID, date); err != nil {
		return nil, err
	}

	return s.repo.Find(ctx, planID)
}

// Current returns the progress of each athlete's current plan that is tracked. "Current" is the plan
// module's own definition, so a next plan written early takes over only on its start date.
func (s *Service) Current(ctx context.Context, coachID uuid.UUID) ([]PlanCount, error) {
	current, err := s.plans.Current(ctx, coachID)
	if err != nil {
		return nil, err
	}

	athleteOf := make(map[uuid.UUID]uuid.UUID, len(current.Plans))
	ids := make([]uuid.UUID, 0, len(current.Plans))
	for _, p := range current.Plans {
		athleteOf[p.ID] = p.AthleteID
		ids = append(ids, p.ID)
	}

	counts, err := s.repo.Counts(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range counts {
		counts[i].AthleteID = athleteOf[counts[i].PlanID]
	}

	return counts, nil
}

// visitDate takes now as a parameter so tests can pin the window: the plan's start date through tomorrow in UTC.
func visitDate(raw string, start, now time.Time) (time.Time, error) {
	y, m, d := now.UTC().Date()
	latest := time.Date(y, m, d+maxDaysAhead, 0, 0, 0, 0, time.UTC)

	raw = strings.TrimSpace(raw)
	date, err := time.Parse(dateLayout, raw)

	var msg string
	switch {
	case raw == "":
		msg = httpx.MsgFieldRequired
	case err != nil:
		msg = httpx.MsgDateFormat
	case date.Before(start) || date.After(latest):
		msg = fmt.Sprintf(msgDateRange, start.Format(dateLayout), latest.Format(dateLayout))
	default:
		return date, nil
	}

	return time.Time{}, &ValidationError{Fields: map[string]string{fieldDate: msg}}
}
