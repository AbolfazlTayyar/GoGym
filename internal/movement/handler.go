package movement

import (
	"errors"
	"net/http"
	"time"

	"github.com/AbolfazlTayyar/gogym/internal/httpx"
	"github.com/AbolfazlTayyar/gogym/internal/tenant"
	"github.com/gin-gonic/gin"
)

const (
	paramID    = "id"
	paramQuery = "q"
)

const msgUniversalReadOnly = "universal movements can't be edited or deleted"

// movementResponse's coach_id is null for a universal movement, which is how a client knows it's read-only.
type movementResponse struct {
	ID          string  `json:"id" example:"e3935578-9941-45f9-b273-4ce0086cb0b8"`
	CoachID     *string `json:"coach_id" example:"3f0b1c6e-2a1d-4f7b-9c3e-6d5a4b3c2d1e"`
	Name        string  `json:"name" example:"Back squat"`
	Category    *string `json:"category" example:"strength"`
	Description *string `json:"description" example:"High bar, below parallel"`
	MuscleGroup *string `json:"muscle_group" example:"legs" enums:"chest,back,shoulders,arms,legs,core,full_body"`
	Equipment   *string `json:"equipment" example:"barbell" enums:"bodyweight,barbell,dumbbell,kettlebell,machine,cable,band,other"`
	MediaURL    *string `json:"media_url" example:"https://example.com/back-squat.mp4"`
	CreatedAt   string  `json:"created_at" example:"2026-10-08T09:30:00Z"`
	UpdatedAt   string  `json:"updated_at" example:"2026-10-08T09:30:00Z"`
}

func newMovementResponse(m *Movement) movementResponse {
	var coachID *string
	if m.CoachID != nil {
		id := m.CoachID.String()
		coachID = &id
	}

	return movementResponse{
		ID:          m.ID.String(),
		CoachID:     coachID,
		Name:        m.Name,
		Category:    m.Category,
		Description: m.Description,
		MuscleGroup: m.MuscleGroup,
		Equipment:   m.Equipment,
		MediaURL:    m.MediaURL,
		CreatedAt:   m.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   m.UpdatedAt.Format(time.RFC3339),
	}
}

// movementRequest serves both create and update. It has no binding tags on purpose: the service owns
// every rule, presence included.
type movementRequest struct {
	Name        string  `json:"name" example:"Bulgarian split squat"`
	Category    *string `json:"category" example:"strength"`
	Description *string `json:"description" example:"Rear foot on a bench"`
	MuscleGroup *string `json:"muscle_group" example:"legs" enums:"chest,back,shoulders,arms,legs,core,full_body"`
	Equipment   *string `json:"equipment" example:"dumbbell" enums:"bodyweight,barbell,dumbbell,kettlebell,machine,cable,band,other"`
}

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// List returns the movements the coach can pick from: their own and the universal ones.
//
//	@Summary		List movements
//	@Description	Lists the authenticated coach's own movements together with the universal, system-seeded ones, ordered by name. A universal movement has a null coach_id and is read-only. Filters combine with AND; an unknown muscle_group or equipment value is a 400, not an empty list. Not paginated.
//	@Tags			movements
//	@Produce		json
//	@Security		BearerAuth
//	@Param			q				query		string	false	"Case-insensitive partial match on name"
//	@Param			muscle_group	query		string	false	"Only movements for this primary muscle group"	Enums(chest, back, shoulders, arms, legs, core, full_body)
//	@Param			equipment		query		string	false	"Only movements using this equipment"			Enums(bodyweight, barbell, dumbbell, kettlebell, machine, cable, band, other)
//	@Success		200				{object}	httpx.SuccessEnvelope{data=[]movementResponse}
//	@Failure		400				{object}	httpx.ErrorEnvelope
//	@Failure		401				{object}	httpx.ErrorEnvelope
//	@Failure		500				{object}	httpx.ErrorEnvelope
//	@Router			/api/v1/movements [get]
func (h *Handler) List(c *gin.Context) {
	coachID, ok := tenant.CoachIDFromContext(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, httpx.CodeUnauthorized, httpx.MsgUnauthorized)
		return
	}

	movements, err := h.svc.List(c.Request.Context(), coachID, ListOptions{
		Query:       c.Query(paramQuery),
		MuscleGroup: c.Query(fieldMuscleGroup),
		Equipment:   c.Query(fieldEquipment),
	})
	if err != nil {
		writeServiceError(c, err)
		return
	}

	items := make([]movementResponse, 0, len(movements))
	for i := range movements {
		items = append(items, newMovementResponse(&movements[i]))
	}

	httpx.OK(c, items)
}

// Create adds a movement to the authenticated coach's own library.
//
//	@Summary		Create a movement
//	@Description	Adds a movement owned by the authenticated coach. The owner comes from the bearer token — a coach_id in the body is ignored. name is required (at most 100 characters); category (at most 50), description (at most 2000), muscle_group and equipment are optional. muscle_group and equipment are matched case-insensitively and stored lowercase.
//	@Tags			movements
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		movementRequest	true	"Movement details"
//	@Success		201		{object}	httpx.SuccessEnvelope{data=movementResponse}
//	@Failure		400		{object}	httpx.ErrorEnvelope
//	@Failure		401		{object}	httpx.ErrorEnvelope
//	@Failure		500		{object}	httpx.ErrorEnvelope
//	@Router			/api/v1/movements [post]
func (h *Handler) Create(c *gin.Context) {
	coachID, ok := tenant.CoachIDFromContext(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, httpx.CodeUnauthorized, httpx.MsgUnauthorized)
		return
	}

	var req movementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ErrorFields(c, http.StatusBadRequest, httpx.CodeValidationFailed, httpx.MsgInvalidRequest, httpx.ValidationFields(err))
		return
	}

	created, err := h.svc.Create(c.Request.Context(), coachID, Input(req))
	if err != nil {
		writeServiceError(c, err)
		return
	}

	httpx.Created(c, newMovementResponse(created))
}

// Update replaces one of the coach's own movements.
//
//	@Summary		Update a movement
//	@Description	Replaces every editable field of one of the authenticated coach's own movements, under the same rules as create: an optional field left out is cleared. media_url is not editable and is kept. A universal movement is a 403; another coach's movement is reported as not found, the same as one that doesn't exist.
//	@Tags			movements
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string			true	"Movement ID"	format(uuid)
//	@Param			request	body		movementRequest	true	"Movement details"
//	@Success		200		{object}	httpx.SuccessEnvelope{data=movementResponse}
//	@Failure		400		{object}	httpx.ErrorEnvelope
//	@Failure		401		{object}	httpx.ErrorEnvelope
//	@Failure		403		{object}	httpx.ErrorEnvelope
//	@Failure		404		{object}	httpx.ErrorEnvelope
//	@Failure		500		{object}	httpx.ErrorEnvelope
//	@Router			/api/v1/movements/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	coachID, ok := tenant.CoachIDFromContext(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, httpx.CodeUnauthorized, httpx.MsgUnauthorized)
		return
	}

	id, ok := httpx.PathUUID(c, paramID)
	if !ok {
		return
	}

	var req movementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ErrorFields(c, http.StatusBadRequest, httpx.CodeValidationFailed, httpx.MsgInvalidRequest, httpx.ValidationFields(err))
		return
	}

	updated, err := h.svc.Update(c.Request.Context(), coachID, id, Input(req))
	if err != nil {
		writeServiceError(c, err)
		return
	}

	httpx.OK(c, newMovementResponse(updated))
}

// Delete removes one of the coach's own movements from their library.
//
//	@Summary		Delete a movement
//	@Description	Removes one of the authenticated coach's own movements from the library; data is null on success. Plans that already use it are left intact and keep showing it, but it can no longer be added to a block. A universal movement is a 403; another coach's movement is reported as not found, the same as one that doesn't exist.
//	@Tags			movements
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Movement ID"	format(uuid)
//	@Success		200	{object}	httpx.SuccessEnvelope
//	@Failure		401	{object}	httpx.ErrorEnvelope
//	@Failure		403	{object}	httpx.ErrorEnvelope
//	@Failure		404	{object}	httpx.ErrorEnvelope
//	@Failure		500	{object}	httpx.ErrorEnvelope
//	@Router			/api/v1/movements/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	coachID, ok := tenant.CoachIDFromContext(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, httpx.CodeUnauthorized, httpx.MsgUnauthorized)
		return
	}

	id, ok := httpx.PathUUID(c, paramID)
	if !ok {
		return
	}

	if err := h.svc.Delete(c.Request.Context(), coachID, id); err != nil {
		writeServiceError(c, err)
		return
	}

	httpx.NoContent(c)
}

// writeServiceError hides unexpected errors behind a generic 500; the detail belongs in logs.
func writeServiceError(c *gin.Context, err error) {
	var verr *ValidationError
	if errors.As(err, &verr) {
		httpx.ErrorFields(c, http.StatusBadRequest, httpx.CodeValidationFailed, httpx.MsgInvalidRequest, verr.Fields)
		return
	}

	if errors.Is(err, ErrUniversal) {
		httpx.Error(c, http.StatusForbidden, httpx.CodeForbidden, msgUniversalReadOnly)
		return
	}

	if errors.Is(err, ErrNotFound) {
		httpx.Error(c, http.StatusNotFound, httpx.CodeNotFound, httpx.MsgNotFound)
		return
	}

	httpx.InternalError(c, err)
}
