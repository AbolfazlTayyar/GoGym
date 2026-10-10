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

const (
	paramPlanID  = "id"
	paramDayID   = "id"
	paramBlockID = "id"
)

const msgNoMovements = "at least one movement is required"

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

// The create requests have no binding tags on purpose: the service owns every rule, presence included.

type createRequest struct {
	AthleteID string  `json:"athlete_id" example:"3f0b1c6e-2a1d-4f7b-9c3e-6d5a4b3c2d1e"`
	StartDate string  `json:"start_date" example:"2026-09-15"`
	Title     string  `json:"title" example:"Cut phase 1"`
	Note      *string `json:"note" example:"Deload every 4th week"`
}

type addDayRequest struct {
	Label      string `json:"label" example:"A"`
	OrderIndex *int   `json:"order_index" example:"0"`
}

type addBlockRequest struct {
	OrderIndex  *int    `json:"order_index" example:"1"`
	Sets        *int    `json:"sets" example:"3"`
	RestSeconds *int    `json:"rest_seconds" example:"60"`
	Notes       *string `json:"notes" example:"superset"`
}

// addMovementRequest gives every field an example: Swagger UI pre-fills a missing one with 0, which fails validation.
type addMovementRequest struct {
	MovementID      string  `json:"movement_id" example:"8e9f0a1b-2c3d-4e5f-8a6b-7c8d9e0f1a2b"`
	Reps            *int    `json:"reps" example:"10"`
	DurationSeconds *int    `json:"duration_seconds" example:"30"`
	Load            *string `json:"load" example:"bodyweight"`
	OrderInBlock    *int    `json:"order_in_block" example:"0"`
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
		movements = append(movements, newBlockMovementResponse(&b.Movements[i]))
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

func newBlockMovementResponse(bm *BlockMovement) blockMovementResponse {
	return blockMovementResponse{
		ID:              bm.ID.String(),
		MovementID:      bm.MovementID.String(),
		Name:            bm.Movement.Name,
		Category:        bm.Movement.Category,
		Reps:            bm.Reps,
		DurationSeconds: bm.DurationSeconds,
		Load:            bm.Load,
		OrderInBlock:    bm.OrderInBlock,
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
//	@Description	Returns the plan with its full nested tree in one response: days ordered by order_index (at most 7), each day's blocks ordered by order_index, and each block's movements ordered by order_in_block with the movement library's name and category inlined. A block is the superset unit: a single movement and a superset have the same shape, differing only in the length of movements. Empty levels are empty arrays, never null. A plan whose athlete belongs to another coach is reported as not found.
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

// Create starts an empty plan for one of the authenticated coach's athletes.
//
//	@Summary		Create a plan
//	@Description	Creates an empty plan for the athlete; add days, blocks and movements to it with the builder endpoints. athlete_id, start_date (YYYY-MM-DD) and title are required, note is optional. The response has the plan detail shape with an empty days array. An athlete that belongs to another coach is reported as not found.
//	@Tags			plans
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		createRequest	true	"Plan details"
//	@Success		201		{object}	httpx.SuccessEnvelope{data=planDetailResponse}
//	@Failure		400		{object}	httpx.ErrorEnvelope
//	@Failure		401		{object}	httpx.ErrorEnvelope
//	@Failure		404		{object}	httpx.ErrorEnvelope
//	@Failure		500		{object}	httpx.ErrorEnvelope
//	@Router			/api/v1/plans [post]
func (h *Handler) Create(c *gin.Context) {
	coachID, ok := tenant.CoachIDFromContext(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, httpx.CodeUnauthorized, httpx.MsgUnauthorized)
		return
	}

	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ErrorFields(c, http.StatusBadRequest, httpx.CodeValidationFailed, httpx.MsgInvalidRequest, httpx.ValidationFields(err))
		return
	}

	created, err := h.svc.Create(c.Request.Context(), coachID, CreateInput(req))
	if err != nil {
		writeServiceError(c, err)
		return
	}

	httpx.Created(c, newPlanDetailResponse(created))
}

// AddDay adds a training day to a plan, in one of its seven slots.
//
//	@Summary		Add a day to a plan
//	@Description	Adds a day to the plan. label must be one of A-G or day1-day7. order_index is the day's slot, 0-6: a plan holds at most 7 days, one per slot, so a slot outside 0-6 or one another day already uses is a 400 on order_index. The response is the new day with an empty blocks array. A plan whose athlete belongs to another coach is reported as not found.
//	@Tags			plans
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string			true	"Plan ID"	format(uuid)
//	@Param			request	body		addDayRequest	true	"Day details"
//	@Success		201		{object}	httpx.SuccessEnvelope{data=dayResponse}
//	@Failure		400		{object}	httpx.ErrorEnvelope
//	@Failure		401		{object}	httpx.ErrorEnvelope
//	@Failure		404		{object}	httpx.ErrorEnvelope
//	@Failure		500		{object}	httpx.ErrorEnvelope
//	@Router			/api/v1/plans/{id}/days [post]
func (h *Handler) AddDay(c *gin.Context) {
	coachID, ok := tenant.CoachIDFromContext(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, httpx.CodeUnauthorized, httpx.MsgUnauthorized)
		return
	}

	planID, ok := httpx.PathUUID(c, paramPlanID)
	if !ok {
		return
	}

	var req addDayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ErrorFields(c, http.StatusBadRequest, httpx.CodeValidationFailed, httpx.MsgInvalidRequest, httpx.ValidationFields(err))
		return
	}

	created, err := h.svc.AddDay(c.Request.Context(), coachID, planID, AddDayInput(req))
	if err != nil {
		writeServiceError(c, err)
		return
	}

	httpx.Created(c, newDayResponse(created))
}

// AddBlock adds a block to a day; its movements are added separately, as one batch.
//
//	@Summary		Add a block to a day
//	@Description	Adds a block to the day. order_index (0-999) and sets (1-50) are required; rest_seconds (0-3600) and notes are optional. A block is the superset unit: add its movements, one or several, with the block movements endpoint. The response is the new block with an empty movements array. A day whose plan's athlete belongs to another coach is reported as not found.
//	@Tags			plans
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string			true	"Day ID"	format(uuid)
//	@Param			request	body		addBlockRequest	true	"Block details"
//	@Success		201		{object}	httpx.SuccessEnvelope{data=blockResponse}
//	@Failure		400		{object}	httpx.ErrorEnvelope
//	@Failure		401		{object}	httpx.ErrorEnvelope
//	@Failure		404		{object}	httpx.ErrorEnvelope
//	@Failure		500		{object}	httpx.ErrorEnvelope
//	@Router			/api/v1/days/{id}/blocks [post]
func (h *Handler) AddBlock(c *gin.Context) {
	coachID, ok := tenant.CoachIDFromContext(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, httpx.CodeUnauthorized, httpx.MsgUnauthorized)
		return
	}

	dayID, ok := httpx.PathUUID(c, paramDayID)
	if !ok {
		return
	}

	var req addBlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ErrorFields(c, http.StatusBadRequest, httpx.CodeValidationFailed, httpx.MsgInvalidRequest, httpx.ValidationFields(err))
		return
	}

	created, err := h.svc.AddBlock(c.Request.Context(), coachID, dayID, AddBlockInput(req))
	if err != nil {
		writeServiceError(c, err)
		return
	}

	httpx.Created(c, newBlockResponse(created))
}

// AddMovements adds one or more movements to a block in a single all-or-nothing call, so a superset is never half-created.
//
//	@Summary		Add movements to a block
//	@Description	Adds every movement in the array to the block, or none of them. Send two or more to build a superset in one call. Each entry needs movement_id and order_in_block (0-999); reps (1-1000), duration_seconds (1-14400) and load (free text, e.g. "70kg", "75% 1RM", "RPE 8", "bodyweight") are optional. Every movement_id must be in the coach's library: their own movements or a universal one. A movement that isn't, or any other invalid entry, rejects the whole batch with a 400 whose fields are keyed by array position, e.g. "[1].movement_id". The response lists the created movements with their library name and category. A block whose plan's athlete belongs to another coach is reported as not found.
//	@Tags			plans
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Block ID"	format(uuid)
//	@Param			request	body		[]addMovementRequest	true	"Movements to add, in one batch"
//	@Success		201		{object}	httpx.SuccessEnvelope{data=[]blockMovementResponse}
//	@Failure		400		{object}	httpx.ErrorEnvelope
//	@Failure		401		{object}	httpx.ErrorEnvelope
//	@Failure		404		{object}	httpx.ErrorEnvelope
//	@Failure		500		{object}	httpx.ErrorEnvelope
//	@Router			/api/v1/blocks/{id}/movements [post]
func (h *Handler) AddMovements(c *gin.Context) {
	coachID, ok := tenant.CoachIDFromContext(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, httpx.CodeUnauthorized, httpx.MsgUnauthorized)
		return
	}

	blockID, ok := httpx.PathUUID(c, paramBlockID)
	if !ok {
		return
	}

	var req []addMovementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ErrorFields(c, http.StatusBadRequest, httpx.CodeValidationFailed, httpx.MsgInvalidRequest, httpx.ValidationFields(err))
		return
	}

	inputs := make([]AddMovementInput, 0, len(req))
	for _, r := range req {
		inputs = append(inputs, AddMovementInput(r))
	}

	created, err := h.svc.AddMovements(c.Request.Context(), coachID, blockID, inputs)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	items := make([]blockMovementResponse, 0, len(created))
	for i := range created {
		items = append(items, newBlockMovementResponse(&created[i]))
	}

	httpx.Created(c, items)
}

// writeServiceError hides unexpected errors behind a generic 500; the detail belongs in logs.
func writeServiceError(c *gin.Context, err error) {
	var verr *ValidationError
	if errors.As(err, &verr) {
		httpx.ErrorFields(c, http.StatusBadRequest, httpx.CodeValidationFailed, httpx.MsgInvalidRequest, verr.Fields)
		return
	}

	if errors.Is(err, ErrNoMovements) {
		httpx.Error(c, http.StatusBadRequest, httpx.CodeValidationFailed, msgNoMovements)
		return
	}

	if errors.Is(err, athlete.ErrNotFound) || errors.Is(err, ErrNotFound) {
		httpx.Error(c, http.StatusNotFound, httpx.CodeNotFound, httpx.MsgNotFound)
		return
	}

	httpx.InternalError(c, err)
}
