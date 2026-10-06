package plan

import (
	"context"
	"errors"
	"time"

	"github.com/AbolfazlTayyar/gogym/internal/athlete"
	"github.com/google/uuid"
)

// ListResult's CurrentID is uuid.Nil when the athlete has no plan that has started yet.
type ListResult struct {
	Plans     []Plan
	CurrentID uuid.UUID
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
	p, err := s.repo.FindByID(ctx, planID)
	if err != nil {
		return nil, err
	}

	if _, err := s.athletes.Get(ctx, coachID, p.AthleteID); err != nil {
		if errors.Is(err, athlete.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	days, err := s.repo.LoadDays(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	p.Days = days

	return p, nil
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
