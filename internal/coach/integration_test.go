package coach_test

import (
	"bytes"
	"context"
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
	"gorm.io/gorm"
)

// newTestRouter also returns the service, since coaches are created through it rather than over HTTP.
func newTestRouter(t *testing.T) (*gin.Engine, *coach.Service) {
	t.Helper()

	return newTestRouterWithDB(testutil.NewDB(t))
}

func newTestRouterWithDB(db *gorm.DB) (*gin.Engine, *coach.Service) {
	repo := coach.NewRepository(db)
	svc := coach.NewService(repo, "test-secret", time.Hour)
	handler := coach.NewHandler(svc, repo)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	v1 := router.Group("/api/v1")
	protected := v1.Group("")
	protected.Use(coach.AuthMiddleware(svc))

	coach.RegisterRoutes(v1, protected, handler)

	return router, svc
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

const (
	testPhone    = "09372144430"
	testPassword = "correct-horse-battery-staple"
)

func createTestCoach(t *testing.T, svc *coach.Service, phone string) *coach.Coach {
	t.Helper()

	created, err := svc.Create(context.Background(), coach.CreateInput{
		FirstName: "Ada",
		LastName:  "Lovelace",
		Phone:     phone,
		Password:  testPassword,
	})
	require.NoError(t, err)

	return created
}

func TestCreateLoginProtectedRoute(t *testing.T) {
	router, svc := newTestRouter(t)

	created := createTestCoach(t, svc, testPhone)

	loginRec := postJSON(t, router, "/api/v1/auth/login", map[string]string{
		"phone":    testPhone,
		"password": testPassword,
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
	require.Equal(t, created.ID.String(), meResp.ID)
	require.Equal(t, testPhone, meResp.Phone)
	require.NotContains(t, meRec.Body.String(), "password_hash")
}

func TestSignupIsNotPublic(t *testing.T) {
	router, _ := newTestRouter(t)

	rec := postJSON(t, router, "/api/v1/auth/signup", map[string]string{
		"first_name": "Ada",
		"last_name":  "Lovelace",
		"phone":      testPhone,
		"password":   testPassword,
	})

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestLogin_WrongPassword(t *testing.T) {
	router, svc := newTestRouter(t)

	createTestCoach(t, svc, testPhone)

	loginRec := postJSON(t, router, "/api/v1/auth/login", map[string]string{
		"phone":    testPhone,
		"password": "wrong-password",
	})

	decodeError(t, loginRec, http.StatusUnauthorized, httpx.CodeInvalidCredentials)
}

func TestLogin_UnknownPhone(t *testing.T) {
	router, _ := newTestRouter(t)

	loginRec := postJSON(t, router, "/api/v1/auth/login", map[string]string{
		"phone":    testPhone,
		"password": testPassword,
	})

	decodeError(t, loginRec, http.StatusUnauthorized, httpx.CodeInvalidCredentials)
}

func TestLogin_DatabaseFailureIsInternalError(t *testing.T) {
	db := testutil.NewDB(t)
	router, _ := newTestRouterWithDB(db)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	loginRec := postJSON(t, router, "/api/v1/auth/login", map[string]string{
		"phone":    testPhone,
		"password": testPassword,
	})

	decodeError(t, loginRec, http.StatusInternalServerError, httpx.CodeInternalError)
	assert.NotContains(t, loginRec.Body.String(), "closed", "driver errors stay out of the response")
}

func TestCreate_DuplicatePhone(t *testing.T) {
	_, svc := newTestRouter(t)

	createTestCoach(t, svc, testPhone)

	_, err := svc.Create(context.Background(), coach.CreateInput{
		FirstName: "Grace",
		LastName:  "Hopper",
		Phone:     testPhone,
		Password:  testPassword,
	})
	require.ErrorIs(t, err, coach.ErrPhoneTaken)
}

func TestResetPassword(t *testing.T) {
	router, svc := newTestRouter(t)

	createTestCoach(t, svc, testPhone)

	const newPassword = "battery-staple-correct-horse"
	require.NoError(t, svc.ResetPassword(context.Background(), testPhone, newPassword))

	oldRec := postJSON(t, router, "/api/v1/auth/login", map[string]string{"phone": testPhone, "password": testPassword})
	decodeError(t, oldRec, http.StatusUnauthorized, httpx.CodeInvalidCredentials)

	var loginResp struct {
		Token string `json:"token"`
	}
	newRec := postJSON(t, router, "/api/v1/auth/login", map[string]string{"phone": testPhone, "password": newPassword})
	decodeSuccess(t, newRec, http.StatusOK, &loginResp)
	require.NotEmpty(t, loginResp.Token)
}

func TestResetPassword_UnknownPhone(t *testing.T) {
	_, svc := newTestRouter(t)

	err := svc.ResetPassword(context.Background(), "09120000000", testPassword)
	require.ErrorIs(t, err, coach.ErrNotFound)
}

// rateLimitProbeAttempts exceeds authRateLimitMax, which this external test package can't reference.
const rateLimitProbeAttempts = 20

// TestAuthRateLimit_Envelope needs a real database: the requests before the 429 are real failed logins.
func TestAuthRateLimit_Envelope(t *testing.T) {
	router, _ := newTestRouter(t)

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
