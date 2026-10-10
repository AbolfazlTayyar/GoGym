package athlete

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/AbolfazlTayyar/gogym/internal/httpx"
	"github.com/AbolfazlTayyar/gogym/internal/tenant"
	"github.com/gin-gonic/gin"
)

// paramID is also the wildcard nested routes like /athletes/:id/measurements must reuse; gin panics on a mismatch.
const paramID = "id"

const (
	paramQuery       = "q"
	paramAthleteType = "athlete_type"
	paramLimit       = "limit"
	paramOffset      = "offset"
)

const msgMustBeInteger = "must be an integer"

// athleteResponse uses pointers for nullable columns so unset ones serialize as null, not a zero value.
type athleteResponse struct {
	ID              string   `json:"id" example:"3f0b1c6e-2a1d-4f7b-9c3e-6d5a4b3c2d1e"`
	FirstName       string   `json:"first_name" example:"Sara"`
	LastName        string   `json:"last_name" example:"Ahmadi"`
	Phone           string   `json:"phone" example:"09372144430"`
	ExperienceLevel *string  `json:"experience_level" example:"beginner"`
	Injuries        *string  `json:"injuries" example:"left shoulder impingement"`
	Goal            *string  `json:"goal" example:"lose 5kg"`
	Height          *float64 `json:"height" example:"172"`
	AthleteType     string   `json:"athlete_type" example:"private"`
	CreatedAt       string   `json:"created_at" example:"2026-09-24T09:30:00Z"`
	UpdatedAt       string   `json:"updated_at" example:"2026-09-24T09:30:00Z"`
}

func newAthleteResponse(a *Athlete) athleteResponse {
	return athleteResponse{
		ID:              a.ID.String(),
		FirstName:       a.FirstName,
		LastName:        a.LastName,
		Phone:           a.Phone,
		ExperienceLevel: a.ExperienceLevel,
		Injuries:        a.Injuries,
		Goal:            a.Goal,
		Height:          a.Height,
		AthleteType:     a.AthleteType,
		CreatedAt:       a.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       a.UpdatedAt.Format(time.RFC3339),
	}
}

// listMeta lives here until a second paginated endpoint needs it; then move it to httpx rather than copy it.
type listMeta struct {
	Total  int64 `json:"total" example:"42"`
	Limit  int   `json:"limit" example:"20"`
	Offset int   `json:"offset" example:"0"`
}

// createRequest has no binding tags on purpose: the service owns every rule, presence included.
type createRequest struct {
	FirstName       string   `json:"first_name" example:"Sara"`
	LastName        string   `json:"last_name" example:"Ahmadi"`
	Phone           string   `json:"phone" example:"09372144430"`
	ExperienceLevel *string  `json:"experience_level" example:"beginner"`
	Injuries        *string  `json:"injuries" example:"left shoulder impingement"`
	Goal            *string  `json:"goal" example:"lose 5kg"`
	Height          *float64 `json:"height" example:"172"`
	AthleteType     *string  `json:"athlete_type" example:"private"`
}

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Create adds an athlete to the authenticated coach's roster.
//
//	@Summary		Create an athlete
//	@Description	Adds an athlete to the authenticated coach's roster. The owning coach comes from the bearer token — a coach_id in the body is ignored. Omitting athlete_type creates a private athlete.
//	@Tags			athletes
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		createRequest	true	"Athlete details"
//	@Success		201		{object}	httpx.SuccessEnvelope{data=athleteResponse}
//	@Failure		400		{object}	httpx.ErrorEnvelope
//	@Failure		401		{object}	httpx.ErrorEnvelope
//	@Failure		500		{object}	httpx.ErrorEnvelope
//	@Router			/api/v1/athletes [post]
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

	input := CreateInput{
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		Phone:           req.Phone,
		ExperienceLevel: req.ExperienceLevel,
		Injuries:        req.Injuries,
		Goal:            req.Goal,
		Height:          req.Height,
	}
	if req.AthleteType != nil {
		input.AthleteType = *req.AthleteType
	}

	created, err := h.svc.Create(c.Request.Context(), coachID, input)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	httpx.Created(c, newAthleteResponse(created))
}

// List returns the authenticated coach's athletes.
//
//	@Summary		List athletes
//	@Description	Lists the authenticated coach's athletes, newest first. Defaults to private athletes only — the coach's ongoing working list — since public athletes are one-off plan-link deliveries; pass athlete_type to widen that. Paginated: meta carries the total, limit and offset.
//	@Tags			athletes
//	@Produce		json
//	@Security		BearerAuth
//	@Param			q				query		string	false	"Case-insensitive partial match on first or last name"
//	@Param			athlete_type	query		string	false	"Which athletes to include; defaults to private"	Enums(private, public, all)
//	@Param			limit			query		int		false	"Page size, 1-100"									default(20)
//	@Param			offset			query		int		false	"Rows to skip"										default(0)
//	@Success		200				{object}	httpx.SuccessEnvelope{data=[]athleteResponse,meta=listMeta}
//	@Failure		400				{object}	httpx.ErrorEnvelope
//	@Failure		401				{object}	httpx.ErrorEnvelope
//	@Failure		500				{object}	httpx.ErrorEnvelope
//	@Router			/api/v1/athletes [get]
func (h *Handler) List(c *gin.Context) {
	coachID, ok := tenant.CoachIDFromContext(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, httpx.CodeUnauthorized, httpx.MsgUnauthorized)
		return
	}

	opts := ListOptions{
		Query: c.Query(paramQuery),
		Type:  c.Query(paramAthleteType),
	}

	fields := make(map[string]string)
	opts.Limit = intParam(c, paramLimit, fields)
	opts.Offset = intParam(c, paramOffset, fields)
	if len(fields) > 0 {
		httpx.ErrorFields(c, http.StatusBadRequest, httpx.CodeValidationFailed, httpx.MsgInvalidRequest, fields)
		return
	}

	result, err := h.svc.List(c.Request.Context(), coachID, opts)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	items := make([]athleteResponse, 0, len(result.Athletes))
	for i := range result.Athletes {
		items = append(items, newAthleteResponse(&result.Athletes[i]))
	}

	httpx.OKWithMeta(c, items, listMeta{Total: result.Total, Limit: result.Limit, Offset: result.Offset})
}

// Get returns one of the authenticated coach's athletes.
//
//	@Summary		Get an athlete
//	@Description	Returns the athlete's full profile. An athlete that belongs to another coach is reported as not found, the same as one that doesn't exist.
//	@Tags			athletes
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Athlete ID"	format(uuid)
//	@Success		200	{object}	httpx.SuccessEnvelope{data=athleteResponse}
//	@Failure		401	{object}	httpx.ErrorEnvelope
//	@Failure		404	{object}	httpx.ErrorEnvelope
//	@Failure		500	{object}	httpx.ErrorEnvelope
//	@Router			/api/v1/athletes/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	coachID, ok := tenant.CoachIDFromContext(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, httpx.CodeUnauthorized, httpx.MsgUnauthorized)
		return
	}

	id, ok := httpx.PathUUID(c, paramID)
	if !ok {
		return
	}

	found, err := h.svc.Get(c.Request.Context(), coachID, id)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	httpx.OK(c, newAthleteResponse(found))
}

// intParam returns 0 for a missing param, which the service treats as unset.
func intParam(c *gin.Context, name string, fields map[string]string) int {
	raw := c.Query(name)
	if raw == "" {
		return 0
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		fields[name] = msgMustBeInteger
		return 0
	}

	return value
}

// writeServiceError hides unexpected errors behind a generic 500; the detail belongs in logs.
func writeServiceError(c *gin.Context, err error) {
	var verr *ValidationError
	if errors.As(err, &verr) {
		httpx.ErrorFields(c, http.StatusBadRequest, httpx.CodeValidationFailed, httpx.MsgInvalidRequest, verr.Fields)
		return
	}

	if errors.Is(err, ErrNotFound) {
		httpx.Error(c, http.StatusNotFound, httpx.CodeNotFound, httpx.MsgNotFound)
		return
	}

	httpx.InternalError(c, err)
}
