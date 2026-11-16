package membership

import (
	"context"
	"errors"
	"net/http"

	"github.com/AbolfazlTayyar/gogym/internal/plan"
	"github.com/AbolfazlTayyar/gogym/internal/platform/httpx"
	"github.com/AbolfazlTayyar/gogym/internal/platform/tenant"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// paramPlanID must match the plan module's /plans/:id wildcard; gin panics on two names for one segment.
const paramPlanID = "id"

const paramDate = "date"

type attendanceResponse struct {
	Tracked bool `json:"tracked" example:"true"`
	// Sessions is how many visits the plan's payment covers, so the client never hard-codes it.
	Sessions int      `json:"sessions" example:"12"`
	Dates    []string `json:"dates" example:"2026-11-01,2026-11-03"`
}

func newAttendanceResponse(a *Attendance) attendanceResponse {
	dates := make([]string, 0, len(a.Dates))
	for _, d := range a.Dates {
		dates = append(dates, d.Format(dateLayout))
	}
	return attendanceResponse{Tracked: a.Tracked, Sessions: SessionsPerPlan, Dates: dates}
}

type currentResponse struct {
	PlanID    string  `json:"plan_id" example:"7c9e6679-7425-40de-944b-e07fc1f90ae7"`
	AthleteID string  `json:"athlete_id" example:"3f0b1c6e-2a1d-4f7b-9c3e-6d5a4b3c2d1e"`
	Visits    int     `json:"visits" example:"7"`
	Sessions  int     `json:"sessions" example:"12"`
	LastDate  *string `json:"last_date" example:"2026-11-14"`
}

type visitRequest struct {
	Date string `json:"date" example:"2026-11-16"`
}

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Get returns a plan's attendance.
//
//	@Summary		Get a plan's attendance
//	@Description	Says whether the plan is tracked session by session and lists its visit dates, oldest first. An untracked plan answers tracked false with no dates. A plan whose athlete belongs to another coach is reported as not found.
//	@Tags			attendance
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Plan ID"	format(uuid)
//	@Success		200	{object}	httpx.SuccessEnvelope{data=attendanceResponse}
//	@Failure		401	{object}	httpx.UnauthorizedEnvelope
//	@Failure		404	{object}	httpx.NotFoundEnvelope
//	@Failure		500	{object}	httpx.InternalErrorEnvelope
//	@Router			/api/v1/plans/{id}/attendance [get]
func (h *Handler) Get(c *gin.Context) {
	h.planAction(c, h.svc.Get)
}

// Track turns on session tracking for a plan.
//
//	@Summary		Track a plan's sessions
//	@Description	Marks the plan as paid per session, so its visits can be recorded. Tracking a plan that is already tracked keeps its visits.
//	@Tags			attendance
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Plan ID"	format(uuid)
//	@Success		200	{object}	httpx.SuccessEnvelope{data=attendanceResponse}
//	@Failure		401	{object}	httpx.UnauthorizedEnvelope
//	@Failure		404	{object}	httpx.NotFoundEnvelope
//	@Failure		500	{object}	httpx.InternalErrorEnvelope
//	@Router			/api/v1/plans/{id}/attendance [put]
func (h *Handler) Track(c *gin.Context) {
	h.planAction(c, h.svc.Track)
}

// Untrack turns off session tracking for a plan.
//
//	@Summary		Stop tracking a plan's sessions
//	@Description	Turns tracking off and permanently deletes the plan's recorded visits. Stopping an untracked plan does nothing.
//	@Tags			attendance
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Plan ID"	format(uuid)
//	@Success		200	{object}	httpx.SuccessEnvelope{data=attendanceResponse}
//	@Failure		401	{object}	httpx.UnauthorizedEnvelope
//	@Failure		404	{object}	httpx.NotFoundEnvelope
//	@Failure		500	{object}	httpx.InternalErrorEnvelope
//	@Router			/api/v1/plans/{id}/attendance [delete]
func (h *Handler) Untrack(c *gin.Context) {
	h.planAction(c, h.svc.Untrack)
}

// AddVisit records that the athlete came to the gym on a date.
//
//	@Summary		Record a visit
//	@Description	Adds a visit to a tracked plan. date is required (YYYY-MM-DD), no earlier than the plan's start date and no later than tomorrow in UTC, since the phone's today can run ahead of the server's. A date already recorded, or a visit past the plan's sessions, is a validation error on date. An untracked plan is reported as not found.
//	@Tags			attendance
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string			true	"Plan ID"	format(uuid)
//	@Param			request	body		visitRequest	true	"Visit date"
//	@Success		201		{object}	httpx.SuccessEnvelope{data=attendanceResponse}
//	@Failure		400		{object}	httpx.ValidationErrorEnvelope
//	@Failure		401		{object}	httpx.UnauthorizedEnvelope
//	@Failure		404		{object}	httpx.NotFoundEnvelope
//	@Failure		500		{object}	httpx.InternalErrorEnvelope
//	@Router			/api/v1/plans/{id}/attendance/visits [post]
func (h *Handler) AddVisit(c *gin.Context) {
	coachID, ok := tenant.CoachIDFromContext(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, httpx.CodeUnauthorized, httpx.MsgUnauthorized)
		return
	}

	planID, ok := httpx.PathUUID(c, paramPlanID)
	if !ok {
		return
	}

	var req visitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ErrorFields(c, http.StatusBadRequest, httpx.CodeValidationFailed, httpx.MsgInvalidRequest, httpx.ValidationFields(err))
		return
	}

	a, err := h.svc.AddVisit(c.Request.Context(), coachID, planID, req.Date)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	httpx.Created(c, newAttendanceResponse(a))
}

// DeleteVisit removes a visit recorded by mistake.
//
//	@Summary		Delete a visit
//	@Description	Removes the plan's visit on a date (YYYY-MM-DD) and returns what is left. A date with no visit is reported as not found.
//	@Tags			attendance
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string	true	"Plan ID"	format(uuid)
//	@Param			date	path		string	true	"Visit date"	format(date)
//	@Success		200		{object}	httpx.SuccessEnvelope{data=attendanceResponse}
//	@Failure		401		{object}	httpx.UnauthorizedEnvelope
//	@Failure		404		{object}	httpx.NotFoundEnvelope
//	@Failure		500		{object}	httpx.InternalErrorEnvelope
//	@Router			/api/v1/plans/{id}/attendance/visits/{date} [delete]
func (h *Handler) DeleteVisit(c *gin.Context) {
	coachID, ok := tenant.CoachIDFromContext(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, httpx.CodeUnauthorized, httpx.MsgUnauthorized)
		return
	}

	planID, ok := httpx.PathUUID(c, paramPlanID)
	if !ok {
		return
	}

	a, err := h.svc.DeleteVisit(c.Request.Context(), coachID, planID, c.Param(paramDate))
	if err != nil {
		writeServiceError(c, err)
		return
	}

	httpx.OK(c, newAttendanceResponse(a))
}

// Current lists the progress of every athlete whose current plan is tracked.
//
//	@Summary		List current attendance
//	@Description	One row per athlete whose current plan is tracked session by session; athletes without one are left out. The home rows match them by athlete_id. last_date is null before the first visit.
//	@Tags			attendance
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	httpx.SuccessEnvelope{data=[]currentResponse}
//	@Failure		401	{object}	httpx.UnauthorizedEnvelope
//	@Failure		500	{object}	httpx.InternalErrorEnvelope
//	@Router			/api/v1/attendance/current [get]
func (h *Handler) Current(c *gin.Context) {
	coachID, ok := tenant.CoachIDFromContext(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, httpx.CodeUnauthorized, httpx.MsgUnauthorized)
		return
	}

	counts, err := h.svc.Current(c.Request.Context(), coachID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	items := make([]currentResponse, 0, len(counts))
	for _, pc := range counts {
		item := currentResponse{
			PlanID:    pc.PlanID.String(),
			AthleteID: pc.AthleteID.String(),
			Visits:    pc.Visits,
			Sessions:  SessionsPerPlan,
		}
		if pc.LastDate != nil {
			last := pc.LastDate.Format(dateLayout)
			item.LastDate = &last
		}
		items = append(items, item)
	}

	httpx.OK(c, items)
}

// planAction runs Get, Track and Untrack, which differ only in the service call.
func (h *Handler) planAction(c *gin.Context, action func(context.Context, uuid.UUID, uuid.UUID) (*Attendance, error)) {
	coachID, ok := tenant.CoachIDFromContext(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, httpx.CodeUnauthorized, httpx.MsgUnauthorized)
		return
	}

	planID, ok := httpx.PathUUID(c, paramPlanID)
	if !ok {
		return
	}

	a, err := action(c.Request.Context(), coachID, planID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	httpx.OK(c, newAttendanceResponse(a))
}

// writeServiceError hides unexpected errors behind a generic 500; the detail belongs in logs.
func writeServiceError(c *gin.Context, err error) {
	var verr *ValidationError
	if errors.As(err, &verr) {
		httpx.ErrorFields(c, http.StatusBadRequest, httpx.CodeValidationFailed, httpx.MsgInvalidRequest, verr.Fields)
		return
	}

	if errors.Is(err, plan.ErrNotFound) || errors.Is(err, ErrNotTracked) || errors.Is(err, ErrVisitNotFound) {
		httpx.Error(c, http.StatusNotFound, httpx.CodeNotFound, httpx.MsgNotFound)
		return
	}

	httpx.InternalError(c, err)
}
