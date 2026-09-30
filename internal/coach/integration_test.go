package coach_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AbolfazlTayyar/gogym/internal/coach"
	"github.com/AbolfazlTayyar/gogym/internal/httpx"
	"github.com/AbolfazlTayyar/gogym/internal/testutil"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRouter(t *testing.T) *gin.Engine {
	t.Helper()

	db := testutil.NewDB(t)

	repo := coach.NewRepository(db)
	svc := coach.NewService(repo, "test-secret", time.Hour)
	handler := coach.NewHandler(svc, repo)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	v1 := router.Group("/api/v1")
	protected := v1.Group("")
	protected.Use(coach.AuthMiddleware(svc))

	coach.RegisterRoutes(v1, protected, handler)

	return router
}

type envelope struct {
	Success bool             `json:"success"`
	Data    json.RawMessage  `json:"data"`
	Error   *httpx.ErrorBody `json:"error"`
	Meta    json.RawMessage  `json:"meta"`
}

func decodeSuccess(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, out any) {
	t.Helper()

	require.Equal(t, wantStatus, rec.Code, "body: %s", rec.Body.String())

	var env envelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))
	assert.True(t, env.Success)
	assert.Nil(t, env.Error, "error is null on a success envelope")
	assert.Nil(t, env.Meta, "nothing populates meta yet, so the key is absent")
	assert.NotContains(t, rec.Body.String(), `"meta"`)

	require.NoError(t, json.Unmarshal(env.Data, out))
}

func decodeError(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, wantCode string) *httpx.ErrorBody {
	t.Helper()

	require.Equal(t, wantStatus, rec.Code, "body: %s", rec.Body.String())

	var env envelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))
	assert.False(t, env.Success)
	assert.Equal(t, "null", string(env.Data), "data is null on a failure envelope")
	require.NotNil(t, env.Error)
	assert.Equal(t, wantCode, env.Error.Code)
	assert.NotEmpty(t, env.Error.Message)

	return env.Error
}

func postJSON(t *testing.T, router *gin.Engine, path string, body map[string]string) *httptest.ResponseRecorder {
	t.Helper()

	encoded, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(encoded))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec
}

func TestSignupLoginProtectedRoute(t *testing.T) {
	router := newTestRouter(t)

	signupRec := postJSON(t, router, "/api/v1/auth/signup", map[string]string{
		"first_name": "Ada",
		"last_name":  "Lovelace",
		"phone":      "09372144430",
		"password":   "correct-horse-battery-staple",
	})

	var signupResp struct {
		ID    string `json:"id"`
		Phone string `json:"phone"`
	}
	decodeSuccess(t, signupRec, http.StatusCreated, &signupResp)
	require.Equal(t, "09372144430", signupResp.Phone)
	require.NotContains(t, signupRec.Body.String(), "password_hash")

	loginRec := postJSON(t, router, "/api/v1/auth/login", map[string]string{
		"phone":    "09372144430",
		"password": "correct-horse-battery-staple",
	})

	var loginResp struct {
		Token string `json:"token"`
	}
	decodeSuccess(t, loginRec, http.StatusOK, &loginResp)
	require.NotEmpty(t, loginResp.Token)

	unauthReq := httptest.NewRequest(http.MethodGet, "/api/v1/coaches/me", nil)
	unauthRec := httptest.NewRecorder()
	router.ServeHTTP(unauthRec, unauthReq)
	decodeError(t, unauthRec, http.StatusUnauthorized, httpx.CodeUnauthorized)

	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/coaches/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+loginResp.Token)
	meRec := httptest.NewRecorder()
	router.ServeHTTP(meRec, meReq)

	var meResp struct {
		ID    string `json:"id"`
		Phone string `json:"phone"`
	}
	decodeSuccess(t, meRec, http.StatusOK, &meResp)
	require.Equal(t, signupResp.ID, meResp.ID)
}

func TestLogin_WrongPassword(t *testing.T) {
	router := newTestRouter(t)

	postJSON(t, router, "/api/v1/auth/signup", map[string]string{
		"first_name": "Ada",
		"last_name":  "Lovelace",
		"phone":      "09163503284",
		"password":   "correct-horse-battery-staple",
	})

	loginRec := postJSON(t, router, "/api/v1/auth/login", map[string]string{
		"phone":    "09163503284",
		"password": "wrong-password",
	})

	decodeError(t, loginRec, http.StatusUnauthorized, httpx.CodeInvalidCredentials)
}

func TestSignup_RejectsNonIranianPhone(t *testing.T) {
	router := newTestRouter(t)

	rec := postJSON(t, router, "/api/v1/auth/signup", map[string]string{
		"first_name": "Ada",
		"last_name":  "Lovelace",
		"phone":      "+15550001111",
		"password":   "correct-horse-battery-staple",
	})

	errBody := decodeError(t, rec, http.StatusBadRequest, httpx.CodeValidationFailed)
	require.Contains(t, errBody.Fields["phone"], "Iranian mobile number")
}

func TestSignup_DuplicatePhoneConflicts(t *testing.T) {
	router := newTestRouter(t)

	body := map[string]string{
		"first_name": "Ada",
		"last_name":  "Lovelace",
		"phone":      "09121234567",
		"password":   "correct-horse-battery-staple",
	}

	first := postJSON(t, router, "/api/v1/auth/signup", body)
	require.Equal(t, http.StatusCreated, first.Code)

	second := postJSON(t, router, "/api/v1/auth/signup", body)
	decodeError(t, second, http.StatusConflict, httpx.CodeConflict)
}

// rateLimitProbeAttempts exceeds authRateLimitMax, which this external test package can't reference.
const rateLimitProbeAttempts = 20

// TestAuthRateLimit_Envelope needs a real database: the requests before the 429 are real failed logins.
func TestAuthRateLimit_Envelope(t *testing.T) {
	router := newTestRouter(t)

	body := map[string]string{"phone": "09121112233", "password": "wrong-password"}

	var rec *httptest.ResponseRecorder
	for range rateLimitProbeAttempts {
		rec = postJSON(t, router, "/api/v1/auth/login", body)
		if rec.Code == http.StatusTooManyRequests {
			break
		}
	}

	decodeError(t, rec, http.StatusTooManyRequests, httpx.CodeRateLimited)
}
