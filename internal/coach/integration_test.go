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
	"github.com/AbolfazlTayyar/gogym/internal/platform/httpx"
	"github.com/AbolfazlTayyar/gogym/internal/platform/testutil"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const testTokenExpiry = time.Hour

// newTestRouter also returns the service, since coaches are created through it rather than over HTTP.
// Its cookies are Secure, as outside development.
func newTestRouter(t *testing.T) (*gin.Engine, *coach.Service) {
	t.Helper()

	return newTestRouterWithDB(testutil.NewDB(t), true)
}

func newTestRouterWithDB(db *gorm.DB, secureCookie bool) (*gin.Engine, *coach.Service) {
	svc := coach.NewService(coach.NewRepository(db), "test-secret", testTokenExpiry)
	handler := coach.NewHandler(svc, secureCookie)

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
	var loginData any
	decodeSuccess(t, loginRec, http.StatusOK, &loginData)
	assert.Nil(t, loginData, "the token travels only in the httpOnly cookie")
	token := testutil.Cookie(t, loginRec, coach.AuthCookieName)

	unauthRec := getMe(router, nil)
	decodeError(t, unauthRec, http.StatusUnauthorized, httpx.CodeUnauthorized)

	for name, authorize := range map[string]func(*http.Request){
		"cookie only": func(r *http.Request) { r.AddCookie(&http.Cookie{Name: coach.AuthCookieName, Value: token}) },
		"bearer only": func(r *http.Request) { r.Header.Set("Authorization", "Bearer "+token) },
	} {
		t.Run(name, func(t *testing.T) {
			meRec := getMe(router, authorize)

			var meResp struct {
				ID    string `json:"id"`
				Phone string `json:"phone"`
			}
			decodeSuccess(t, meRec, http.StatusOK, &meResp)
			require.Equal(t, created.ID.String(), meResp.ID)
			require.Equal(t, testPhone, meResp.Phone)
			require.NotContains(t, meRec.Body.String(), "password_hash")
		})
	}
}

func getMe(router *gin.Engine, authorize func(*http.Request)) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/coaches/me", nil)
	if authorize != nil {
		authorize(req)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func authCookie(t *testing.T, rec *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()

	for _, c := range rec.Result().Cookies() {
		if c.Name == coach.AuthCookieName {
			return c
		}
	}
	require.Fail(t, "no auth cookie set")
	return nil
}

func TestLogin_SetsHttpOnlyCookie(t *testing.T) {
	router, svc := newTestRouter(t)
	createTestCoach(t, svc, testPhone)

	rec := postJSON(t, router, "/api/v1/auth/login", map[string]string{"phone": testPhone, "password": testPassword})
	require.Equal(t, http.StatusOK, rec.Code)

	cookie := authCookie(t, rec)
	assert.NotEmpty(t, cookie.Value)
	assert.True(t, cookie.HttpOnly)
	assert.True(t, cookie.Secure)
	assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
	assert.Equal(t, "/api/v1", cookie.Path)
	assert.Equal(t, int(testTokenExpiry.Seconds()), cookie.MaxAge)
	assert.NotContains(t, rec.Body.String(), cookie.Value, "the token is not in the body")
}

func TestLogin_CookieNotSecureInDevelopment(t *testing.T) {
	db := testutil.NewDB(t)
	router, svc := newTestRouterWithDB(db, false)
	createTestCoach(t, svc, testPhone)

	rec := postJSON(t, router, "/api/v1/auth/login", map[string]string{"phone": testPhone, "password": testPassword})
	require.Equal(t, http.StatusOK, rec.Code)

	cookie := authCookie(t, rec)
	assert.False(t, cookie.Secure, "local dev is plain http")
	assert.True(t, cookie.HttpOnly)
}

func TestLogout_ClearsCookie(t *testing.T) {
	router, _ := newTestRouter(t)

	rec := postJSON(t, router, "/api/v1/auth/logout", nil)

	var data any
	decodeSuccess(t, rec, http.StatusOK, &data)
	cookie := authCookie(t, rec)
	assert.Empty(t, cookie.Value)
	assert.Negative(t, cookie.MaxAge, "Max-Age=0 tells the browser to delete it")
	assert.Equal(t, "/api/v1", cookie.Path, "the path must match for the browser to delete the same cookie")
	assert.True(t, cookie.HttpOnly)
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
	router, _ := newTestRouterWithDB(db, true)

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

	newRec := postJSON(t, router, "/api/v1/auth/login", map[string]string{"phone": testPhone, "password": newPassword})
	require.Equal(t, http.StatusOK, newRec.Code)
	testutil.Cookie(t, newRec, coach.AuthCookieName)
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
