package coach

import (
	"errors"
	"net/http"

	"github.com/AbolfazlTayyar/gogym/internal/platform/httpx"
	"github.com/AbolfazlTayyar/gogym/internal/platform/tenant"
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

// AuthCookieName carries the JWT for the browser; see docs/adr/0031.
const AuthCookieName = "gogym_auth"

// authCookiePath is the API prefix, so the cookie never rides along on requests for the frontend's files.
const authCookiePath = "/api/v1"

type Handler struct {
	svc          *Service
	ipLimiter    *rateLimiter
	phoneLimiter *rateLimiter
	// secureCookie is off only in development, which is plain http (a phone on the LAN can't use https).
	secureCookie bool
}

func NewHandler(svc *Service, secureCookie bool) *Handler {
	return &Handler{
		svc:          svc,
		ipLimiter:    newRateLimiter(authRateLimitMax, authRateLimitWindow),
		phoneLimiter: newRateLimiter(authRateLimitMax, authRateLimitWindow),
		secureCookie: secureCookie,
	}
}

// setAuthCookie with an empty token and maxAge -1 clears the cookie.
func (h *Handler) setAuthCookie(c *gin.Context, token string, maxAge int) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     AuthCookieName,
		Value:    token,
		Path:     authCookiePath,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   h.secureCookie,
		SameSite: http.SameSiteStrictMode,
	})
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
// @Description	Verifies phone + password and sets the signed JWT as an httpOnly cookie (gogym_auth, Path=/api/v1, SameSite=Strict, Secure outside development) that the browser sends on every later call. The token is not in the body, so page scripts can't read it. The error is the same whether the phone or the password was wrong.
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			request	body		loginRequest	true	"Login credentials"
// @Success		200		{object}	httpx.NoContentEnvelope
// @Header			200		{string}	Set-Cookie	"gogym_auth=<jwt>; Path=/api/v1; Max-Age=<JWT_EXPIRY>; HttpOnly; Secure; SameSite=Strict"
// @Failure		400		{object}	httpx.ValidationErrorEnvelope
// @Failure		401		{object}	httpx.InvalidCredentialsEnvelope
// @Failure		429		{object}	httpx.RateLimitedEnvelope
// @Failure		500		{object}	httpx.InternalErrorEnvelope
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
		if errors.Is(err, ErrInvalidCredentials) {
			httpx.Error(c, http.StatusUnauthorized, httpx.CodeInvalidCredentials, httpx.MsgInvalidCredentials)
			return
		}
		httpx.InternalError(c, err)
		return
	}

	h.setAuthCookie(c, token, int(h.svc.TokenExpiry().Seconds()))
	httpx.NoContent(c)
}

// Logout clears the auth cookie, which page scripts can't do because it is httpOnly.
//
// @Summary		Log out a coach
// @Description	Clears the gogym_auth cookie. Needs no valid token, so an expired session can still sign out. A token sent in the Authorization header stays valid until it expires.
// @Tags			auth
// @Produce		json
// @Success		200	{object}	httpx.NoContentEnvelope
// @Header			200	{string}	Set-Cookie	"gogym_auth=; Path=/api/v1; Max-Age=0; HttpOnly; Secure; SameSite=Strict"
// @Router			/api/v1/auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	h.setAuthCookie(c, "", -1)
	httpx.NoContent(c)
}

// Me returns the authenticated coach's own profile.
//
// @Summary		Get the authenticated coach
// @Description	Returns the profile of the coach identified by the auth cookie or the bearer token.
// @Tags			auth
// @Produce		json
// @Security		BearerAuth
// @Success		200	{object}	httpx.SuccessEnvelope{data=coachResponse}
// @Failure		401	{object}	httpx.UnauthorizedEnvelope
// @Failure		404	{object}	httpx.NotFoundEnvelope
// @Failure		500	{object}	httpx.InternalErrorEnvelope
// @Router			/api/v1/coaches/me [get]
func (h *Handler) Me(c *gin.Context) {
	coachID, ok := tenant.CoachIDFromContext(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, httpx.CodeUnauthorized, httpx.MsgUnauthorized)
		return
	}

	found, err := h.svc.Get(c.Request.Context(), coachID)
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
