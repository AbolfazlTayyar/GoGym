package coach

import (
	"errors"
	"net/http"

	"github.com/AbolfazlTayyar/gogym/internal/httpx"
	"github.com/AbolfazlTayyar/gogym/internal/tenant"
	"github.com/gin-gonic/gin"
)

// coachResponse must never include PasswordHash.
type coachResponse struct {
	ID        string `json:"id" example:"b8f1c9de-8f0a-4c1e-9a2b-2f6b6b8e2b3a"`
	FirstName string `json:"first_name" example:"Ada"`
	LastName  string `json:"last_name" example:"Lovelace"`
	Phone     string `json:"phone" example:"09372144430"`
}

func newCoachResponse(c *Coach) coachResponse {
	return coachResponse{
		ID:        c.ID.String(),
		FirstName: c.FirstName,
		LastName:  c.LastName,
		Phone:     c.Phone,
	}
}

type loginRequest struct {
	Phone    string `json:"phone" binding:"required" example:"09372144430"`
	Password string `json:"password" binding:"required" example:"correct-horse-battery-staple"`
}

type loginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type Handler struct {
	svc          *Service
	repo         *Repository
	ipLimiter    *rateLimiter
	phoneLimiter *rateLimiter
}

func NewHandler(svc *Service, repo *Repository) *Handler {
	return &Handler{
		svc:          svc,
		repo:         repo,
		ipLimiter:    newRateLimiter(authRateLimitMax, authRateLimitWindow),
		phoneLimiter: newRateLimiter(authRateLimitMax, authRateLimitWindow),
	}
}

// rateLimited writes the 429 itself; callers just return when it reports true.
func (h *Handler) rateLimited(c *gin.Context, phone string) bool {
	if !h.ipLimiter.Allow(c.ClientIP()) {
		httpx.Error(c, http.StatusTooManyRequests, httpx.CodeRateLimited, httpx.MsgTooManyRequests)
		return true
	}
	if phone != "" && !h.phoneLimiter.Allow(phone) {
		httpx.Error(c, http.StatusTooManyRequests, httpx.CodeRateLimited, httpx.MsgTooManyRequests)
		return true
	}
	return false
}

// Login handles coach authentication.
//
// @Summary		Log in a coach
// @Description	Verifies phone + password and returns a signed JWT. The error is the same whether the phone or the password was wrong.
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			request	body		loginRequest	true	"Login credentials"
// @Success		200		{object}	httpx.SuccessEnvelope{data=loginResponse}
// @Failure		400		{object}	httpx.ErrorEnvelope
// @Failure		401		{object}	httpx.ErrorEnvelope
// @Failure		429		{object}	httpx.ErrorEnvelope
// @Router			/api/v1/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ErrorFields(c, http.StatusBadRequest, httpx.CodeValidationFailed, httpx.MsgInvalidRequest, httpx.ValidationFields(err))
		return
	}

	if h.rateLimited(c, req.Phone) {
		return
	}

	token, err := h.svc.Login(c.Request.Context(), req.Phone, req.Password)
	if err != nil {
		httpx.Error(c, http.StatusUnauthorized, httpx.CodeInvalidCredentials, "invalid credentials")
		return
	}

	httpx.OK(c, loginResponse{Token: token})
}

// Me returns the authenticated coach's own profile.
//
// @Summary		Get the authenticated coach
// @Description	Returns the profile of the coach identified by the bearer token.
// @Tags			auth
// @Produce		json
// @Security		BearerAuth
// @Success		200	{object}	httpx.SuccessEnvelope{data=coachResponse}
// @Failure		401	{object}	httpx.ErrorEnvelope
// @Failure		404	{object}	httpx.ErrorEnvelope
// @Failure		500	{object}	httpx.ErrorEnvelope
// @Router			/api/v1/coaches/me [get]
func (h *Handler) Me(c *gin.Context) {
	coachID, ok := tenant.CoachIDFromContext(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, httpx.CodeUnauthorized, httpx.MsgUnauthorized)
		return
	}

	found, err := h.repo.FindByID(c.Request.Context(), coachID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.Error(c, http.StatusNotFound, httpx.CodeNotFound, httpx.MsgNotFound)
			return
		}
		httpx.InternalError(c, err)
		return
	}

	httpx.OK(c, newCoachResponse(found))
}
