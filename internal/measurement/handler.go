package measurement

import (
	"errors"
	"net/http"
	"time"

	"github.com/AbolfazlTayyar/gogym/internal/athlete"
	"github.com/AbolfazlTayyar/gogym/internal/platform/httpx"
	"github.com/AbolfazlTayyar/gogym/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

// paramAthleteID must match the athlete module's /athletes/:id wildcard; gin panics on two names for one segment.
const paramAthleteID = "id"

const paramMeasurementID = "id"

// measurementResponse is one chart point; unrecorded metrics are null so a chart can skip them rather than plot 0.
type measurementResponse struct {
	ID        string   `json:"id" example:"7c9e6679-7425-40de-944b-e07fc1f90ae7"`
	AthleteID string   `json:"athlete_id" example:"3f0b1c6e-2a1d-4f7b-9c3e-6d5a4b3c2d1e"`
	Date      string   `json:"date" example:"2026-09-24"`
	Weight    *float64 `json:"weight" example:"72.5"`
	Chest     *float64 `json:"chest" example:"98"`
	Waist     *float64 `json:"waist" example:"81.5"`
	Arm       *float64 `json:"arm" example:"34"`
	Thigh     *float64 `json:"thigh" example:"56"`
	Hip       *float64 `json:"hip" example:"99"`
	CreatedAt string   `json:"created_at" example:"2026-09-24T09:30:00Z"`
	UpdatedAt string   `json:"updated_at" example:"2026-09-24T09:30:00Z"`
}

func newMeasurementResponse(m *Measurement) measurementResponse {
	return measurementResponse{
		ID:        m.ID.String(),
		AthleteID: m.AthleteID.String(),
		Date:      m.Date.Format(dateLayout),
		Weight:    m.Weight,
		Chest:     m.Chest,
		Waist:     m.Waist,
		Arm:       m.Arm,
		Thigh:     m.Thigh,
		Hip:       m.Hip,
		CreatedAt: m.CreatedAt.Format(time.RFC3339),
		UpdatedAt: m.UpdatedAt.Format(time.RFC3339),
	}
}

// measurementRequest serves both create and update. It has no binding tags on purpose: the service owns
// every rule, presence included.
type measurementRequest struct {
	Date   string   `json:"date" example:"2026-09-24"`
	Weight *float64 `json:"weight" example:"72.5"`
	Chest  *float64 `json:"chest" example:"98"`
	Waist  *float64 `json:"waist" example:"81.5"`
	Arm    *float64 `json:"arm" example:"34"`
	Thigh  *float64 `json:"thigh" example:"56"`
	Hip    *float64 `json:"hip" example:"99"`
}

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Create logs a check-in for one of the authenticated coach's athletes.
//
//	@Summary		Add a measurement
//	@Description	Records one check-in for the athlete. date is required (YYYY-MM-DD); every metric is optional, but at least one must be given. Weight is in kilograms, circumferences in centimetres. An athlete that belongs to another coach is reported as not found.
//	@Tags			measurements
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string			true	"Athlete ID"	format(uuid)
//	@Param			request	body		measurementRequest	true	"Measurement values"
//	@Success		201		{object}	httpx.SuccessEnvelope{data=measurementResponse}
//	@Failure		400		{object}	httpx.ValidationErrorEnvelope
//	@Failure		401		{object}	httpx.UnauthorizedEnvelope
//	@Failure		404		{object}	httpx.NotFoundEnvelope
//	@Failure		500		{object}	httpx.InternalErrorEnvelope
//	@Router			/api/v1/athletes/{id}/measurements [post]
func (h *Handler) Create(c *gin.Context) {
	coachID, ok := tenant.CoachIDFromContext(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, httpx.CodeUnauthorized, httpx.MsgUnauthorized)
		return
	}

	athleteID, ok := httpx.PathUUID(c, paramAthleteID)
	if !ok {
		return
	}

	var req measurementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ErrorFields(c, http.StatusBadRequest, httpx.CodeValidationFailed, httpx.MsgInvalidRequest, httpx.ValidationFields(err))
		return
	}

	created, err := h.svc.Create(c.Request.Context(), coachID, athleteID, Input(req))
	if err != nil {
		writeServiceError(c, err)
		return
	}

	httpx.Created(c, newMeasurementResponse(created))
}

// List returns an athlete's measurement history as chart-ready points.
//
//	@Summary		List measurements
//	@Description	Returns every measurement for the athlete, oldest first, ready to plot as a trend chart. Not paginated: one check-in per visit stays small. An athlete with no measurements yields an empty array; an athlete that belongs to another coach is reported as not found.
//	@Tags			measurements
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Athlete ID"	format(uuid)
//	@Success		200	{object}	httpx.SuccessEnvelope{data=[]measurementResponse}
//	@Failure		401	{object}	httpx.UnauthorizedEnvelope
//	@Failure		404	{object}	httpx.NotFoundEnvelope
//	@Failure		500	{object}	httpx.InternalErrorEnvelope
//	@Router			/api/v1/athletes/{id}/measurements [get]
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

	measurements, err := h.svc.List(c.Request.Context(), coachID, athleteID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	items := make([]measurementResponse, 0, len(measurements))
	for i := range measurements {
		items = append(items, newMeasurementResponse(&measurements[i]))
	}

	httpx.OK(c, items)
}

// Update corrects one of the coach's measurements.
//
//	@Summary		Update a measurement
//	@Description	Replaces the date and every metric of a measurement, under the same rules as create: a metric left out is cleared, and at least one must be given. A measurement whose athlete belongs to another coach is reported as not found, the same as one that doesn't exist.
//	@Tags			measurements
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string				true	"Measurement ID"	format(uuid)
//	@Param			request	body		measurementRequest	true	"Measurement values"
//	@Success		200		{object}	httpx.SuccessEnvelope{data=measurementResponse}
//	@Failure		400		{object}	httpx.ValidationErrorEnvelope
//	@Failure		401		{object}	httpx.UnauthorizedEnvelope
//	@Failure		404		{object}	httpx.NotFoundEnvelope
//	@Failure		500		{object}	httpx.InternalErrorEnvelope
//	@Router			/api/v1/measurements/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	coachID, ok := tenant.CoachIDFromContext(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, httpx.CodeUnauthorized, httpx.MsgUnauthorized)
		return
	}

	id, ok := httpx.PathUUID(c, paramMeasurementID)
	if !ok {
		return
	}

	var req measurementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ErrorFields(c, http.StatusBadRequest, httpx.CodeValidationFailed, httpx.MsgInvalidRequest, httpx.ValidationFields(err))
		return
	}

	updated, err := h.svc.Update(c.Request.Context(), coachID, id, Input(req))
	if err != nil {
		writeServiceError(c, err)
		return
	}

	httpx.OK(c, newMeasurementResponse(updated))
}

// Delete removes one of the coach's measurements for good.
//
//	@Summary		Delete a measurement
//	@Description	Permanently removes a measurement from the athlete's history; data is null on success. There is no undo. A measurement whose athlete belongs to another coach is reported as not found, the same as one that doesn't exist.
//	@Tags			measurements
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Measurement ID"	format(uuid)
//	@Success		200	{object}	httpx.NoContentEnvelope
//	@Failure		401	{object}	httpx.UnauthorizedEnvelope
//	@Failure		404	{object}	httpx.NotFoundEnvelope
//	@Failure		500	{object}	httpx.InternalErrorEnvelope
//	@Router			/api/v1/measurements/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	coachID, ok := tenant.CoachIDFromContext(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, httpx.CodeUnauthorized, httpx.MsgUnauthorized)
		return
	}

	id, ok := httpx.PathUUID(c, paramMeasurementID)
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

	if errors.Is(err, athlete.ErrNotFound) || errors.Is(err, ErrNotFound) {
		httpx.Error(c, http.StatusNotFound, httpx.CodeNotFound, httpx.MsgNotFound)
		return
	}

	httpx.InternalError(c, err)
}
