package plan

import (
	"errors"
	"net/http"
	"time"

	"github.com/AbolfazlTayyar/gogym/internal/athlete"
	"github.com/AbolfazlTayyar/gogym/internal/httpx"
	"github.com/AbolfazlTayyar/gogym/internal/tenant"
	"github.com/gin-gonic/gin"
)

// paramAthleteID must match the athlete module's /athletes/:id wildcard; gin panics on two names for one segment.
const paramAthleteID = "id"

type planSummaryResponse struct {
	ID        string  `json:"id" example:"9b2d4f1a-6c3e-4a7b-8d5f-1e2c3b4a5d6e"`
	AthleteID string  `json:"athlete_id" example:"3f0b1c6e-2a1d-4f7b-9c3e-6d5a4b3c2d1e"`
	Title     string  `json:"title" example:"Cut phase 1"`
	Note      *string `json:"note" example:"Deload every 4th week"`
	StartDate string  `json:"start_date" example:"2026-09-15"`
	IsCurrent bool    `json:"is_current" example:"true"`
	CreatedAt string  `json:"created_at" example:"2026-09-10T18:00:00Z"`
	UpdatedAt string  `json:"updated_at" example:"2026-09-10T18:00:00Z"`
}

func newPlanSummaryResponse(p *Plan, isCurrent bool) planSummaryResponse {
	return planSummaryResponse{
		ID:        p.ID.String(),
		AthleteID: p.AthleteID.String(),
		Title:     p.Title,
		Note:      p.Note,
		StartDate: p.StartDate.Format(time.DateOnly),
		IsCurrent: isCurrent,
		CreatedAt: p.CreatedAt.Format(time.RFC3339),
		UpdatedAt: p.UpdatedAt.Format(time.RFC3339),
	}
}

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// List returns an athlete's plans, newest first, with the one in effect today marked.
//
//	@Summary		List an athlete's plans
//	@Description	Returns every plan for the athlete, newest start_date first. At most one plan has is_current true: the latest one whose start_date is today or earlier. A plan starting in the future is upcoming, not current, and if every plan starts in the future none is current. Not paginated. An athlete with no plans yields an empty array; an athlete that belongs to another coach is reported as not found.
//	@Tags			plans
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Athlete ID"	format(uuid)
//	@Success		200	{object}	httpx.SuccessEnvelope{data=[]planSummaryResponse}
//	@Failure		401	{object}	httpx.ErrorEnvelope
//	@Failure		404	{object}	httpx.ErrorEnvelope
//	@Failure		500	{object}	httpx.ErrorEnvelope
//	@Router			/api/v1/athletes/{id}/plans [get]
func (h *Handler) List(c *gin.Context) {
	coachID, ok := tenant.CoachIDFromContext(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, httpx.CodeUnauthorized, httpx.MsgUnauthorized)
		return
	}

	athleteID, ok := httpx.PathUUID(c, paramAthleteID)
	if !ok {
		return
	}

	result, err := h.svc.List(c.Request.Context(), coachID, athleteID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	items := make([]planSummaryResponse, 0, len(result.Plans))
	for i := range result.Plans {
		p := &result.Plans[i]
		items = append(items, newPlanSummaryResponse(p, p.ID == result.CurrentID))
	}

	httpx.OK(c, items)
}

// writeServiceError hides unexpected errors behind a generic 500; the detail belongs in logs.
func writeServiceError(c *gin.Context, err error) {
	if errors.Is(err, athlete.ErrNotFound) {
		httpx.Error(c, http.StatusNotFound, httpx.CodeNotFound, httpx.MsgNotFound)
		return
	}

	httpx.Error(c, http.StatusInternalServerError, httpx.CodeInternalError, httpx.MsgInternalError)
}
