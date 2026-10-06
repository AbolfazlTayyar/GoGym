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

const paramPlanID = "id"

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

// planDetailResponse nests every level as an array, never null, so an empty day or plan needs no special case.
type planDetailResponse struct {
	ID        string        `json:"id" example:"9b2d4f1a-6c3e-4a7b-8d5f-1e2c3b4a5d6e"`
	AthleteID string        `json:"athlete_id" example:"3f0b1c6e-2a1d-4f7b-9c3e-6d5a4b3c2d1e"`
	Title     string        `json:"title" example:"Cut phase 1"`
	Note      *string       `json:"note" example:"Deload every 4th week"`
	StartDate string        `json:"start_date" example:"2026-09-15"`
	CreatedAt string        `json:"created_at" example:"2026-09-10T18:00:00Z"`
	UpdatedAt string        `json:"updated_at" example:"2026-09-10T18:00:00Z"`
	Days      []dayResponse `json:"days"`
}

type dayResponse struct {
	ID         string          `json:"id" example:"5c1e2d3f-4a5b-4c6d-8e7f-9a0b1c2d3e4f"`
	Label      string          `json:"label" example:"A"`
	OrderIndex int             `json:"order_index" example:"0"`
	Blocks     []blockResponse `json:"blocks"`
}

// blockResponse is the superset unit: a single movement and a superset differ only in len(movements).
type blockResponse struct {
	ID          string                  `json:"id" example:"7d2e3f4a-5b6c-4d7e-8f9a-0b1c2d3e4f5a"`
	OrderIndex  int                     `json:"order_index" example:"1"`
	Sets        int                     `json:"sets" example:"3"`
	RestSeconds *int                    `json:"rest_seconds" example:"60"`
	Notes       *string                 `json:"notes" example:"superset"`
	Movements   []blockMovementResponse `json:"movements"`
}

// blockMovementResponse carries the library's name and category so the client needs no lookup per movement.
type blockMovementResponse struct {
	ID              string  `json:"id" example:"2a3b4c5d-6e7f-4a8b-9c0d-1e2f3a4b5c6d"`
	MovementID      string  `json:"movement_id" example:"8e9f0a1b-2c3d-4e5f-8a6b-7c8d9e0f1a2b"`
	Name            string  `json:"name" example:"Pull-up"`
	Category        *string `json:"category" example:"strength"`
	Reps            *int    `json:"reps" example:"10"`
	DurationSeconds *int    `json:"duration_seconds"`
	Load            *string `json:"load" example:"bodyweight"`
	OrderInBlock    int     `json:"order_in_block" example:"0"`
}

func newPlanDetailResponse(p *Plan) planDetailResponse {
	days := make([]dayResponse, 0, len(p.Days))
	for i := range p.Days {
		days = append(days, newDayResponse(&p.Days[i]))
	}

	return planDetailResponse{
		ID:        p.ID.String(),
		AthleteID: p.AthleteID.String(),
		Title:     p.Title,
		Note:      p.Note,
		StartDate: p.StartDate.Format(time.DateOnly),
		CreatedAt: p.CreatedAt.Format(time.RFC3339),
		UpdatedAt: p.UpdatedAt.Format(time.RFC3339),
		Days:      days,
	}
}

func newDayResponse(d *Day) dayResponse {
	blocks := make([]blockResponse, 0, len(d.Blocks))
	for i := range d.Blocks {
		blocks = append(blocks, newBlockResponse(&d.Blocks[i]))
	}

	return dayResponse{
		ID:         d.ID.String(),
		Label:      d.Label,
		OrderIndex: d.OrderIndex,
		Blocks:     blocks,
	}
}

func newBlockResponse(b *Block) blockResponse {
	movements := make([]blockMovementResponse, 0, len(b.Movements))
	for i := range b.Movements {
		bm := &b.Movements[i]
		movements = append(movements, blockMovementResponse{
			ID:              bm.ID.String(),
			MovementID:      bm.MovementID.String(),
			Name:            bm.Movement.Name,
			Category:        bm.Movement.Category,
			Reps:            bm.Reps,
			DurationSeconds: bm.DurationSeconds,
			Load:            bm.Load,
			OrderInBlock:    bm.OrderInBlock,
		})
	}

	return blockResponse{
		ID:          b.ID.String(),
		OrderIndex:  b.OrderIndex,
		Sets:        b.Sets,
		RestSeconds: b.RestSeconds,
		Notes:       b.Notes,
		Movements:   movements,
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

// Get returns one plan with its whole day → block → movement tree, ordered for display.
//
//	@Summary		Get a plan with its days, blocks and movements
//	@Description	Returns the plan with its full nested tree in one response: days ordered by order_index, each day's blocks ordered by order_index, and each block's movements ordered by order_in_block with the movement library's name and category inlined. A block is the superset unit: a single movement and a superset have the same shape, differing only in the length of movements. Empty levels are empty arrays, never null. A plan whose athlete belongs to another coach is reported as not found.
//	@Tags			plans
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Plan ID"	format(uuid)
//	@Success		200	{object}	httpx.SuccessEnvelope{data=planDetailResponse}
//	@Failure		401	{object}	httpx.ErrorEnvelope
//	@Failure		404	{object}	httpx.ErrorEnvelope
//	@Failure		500	{object}	httpx.ErrorEnvelope
//	@Router			/api/v1/plans/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	coachID, ok := tenant.CoachIDFromContext(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, httpx.CodeUnauthorized, httpx.MsgUnauthorized)
		return
	}

	planID, ok := httpx.PathUUID(c, paramPlanID)
	if !ok {
		return
	}

	p, err := h.svc.Get(c.Request.Context(), coachID, planID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	httpx.OK(c, newPlanDetailResponse(p))
}

// writeServiceError hides unexpected errors behind a generic 500; the detail belongs in logs.
func writeServiceError(c *gin.Context, err error) {
	if errors.Is(err, athlete.ErrNotFound) || errors.Is(err, ErrNotFound) {
		httpx.Error(c, http.StatusNotFound, httpx.CodeNotFound, httpx.MsgNotFound)
		return
	}

	httpx.Error(c, http.StatusInternalServerError, httpx.CodeInternalError, httpx.MsgInternalError)
}
