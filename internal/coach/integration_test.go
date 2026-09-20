package coach_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AbolfazlTayyar/gogym/internal/coach"
	"github.com/AbolfazlTayyar/gogym/internal/testutil"
	"github.com/gin-gonic/gin"
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

func TestSignupLoginProtectedRoute(t *testing.T) {
	router := newTestRouter(t)

	signupBody, err := json.Marshal(map[string]string{
		"first_name": "Ada",
		"last_name":  "Lovelace",
		"phone":      "09372144430",
		"password":   "correct-horse-battery-staple",
	})
	require.NoError(t, err)

	signupReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(signupBody))
	signupReq.Header.Set("Content-Type", "application/json")
	signupRec := httptest.NewRecorder()
	router.ServeHTTP(signupRec, signupReq)
	require.Equal(t, http.StatusCreated, signupRec.Code)

	var signupResp struct {
		ID    string `json:"id"`
		Phone string `json:"phone"`
	}
	require.NoError(t, json.Unmarshal(signupRec.Body.Bytes(), &signupResp))
	require.Equal(t, "09372144430", signupResp.Phone)
	require.NotContains(t, signupRec.Body.String(), "password_hash")

	loginBody, err := json.Marshal(map[string]string{
		"phone":    "09372144430",
		"password": "correct-horse-battery-staple",
	})
	require.NoError(t, err)

	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)
	require.Equal(t, http.StatusOK, loginRec.Code)

	var loginResp struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.Unmarshal(loginRec.Body.Bytes(), &loginResp))
	require.NotEmpty(t, loginResp.Token)

	// Without a token, a protected route is rejected.
	unauthReq := httptest.NewRequest(http.MethodGet, "/api/v1/coaches/me", nil)
	unauthRec := httptest.NewRecorder()
	router.ServeHTTP(unauthRec, unauthReq)
	require.Equal(t, http.StatusUnauthorized, unauthRec.Code)

	// With the token, the protected route succeeds.
	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/coaches/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+loginResp.Token)
	meRec := httptest.NewRecorder()
	router.ServeHTTP(meRec, meReq)
	require.Equal(t, http.StatusOK, meRec.Code)

	var meResp struct {
		ID    string `json:"id"`
		Phone string `json:"phone"`
	}
	require.NoError(t, json.Unmarshal(meRec.Body.Bytes(), &meResp))
	require.Equal(t, signupResp.ID, meResp.ID)
}

func TestLogin_WrongPassword(t *testing.T) {
	router := newTestRouter(t)

	signupBody, _ := json.Marshal(map[string]string{
		"first_name": "Ada",
		"last_name":  "Lovelace",
		"phone":      "09163503284",
		"password":   "correct-horse-battery-staple",
	})
	signupReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(signupBody))
	signupReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(httptest.NewRecorder(), signupReq)

	loginBody, _ := json.Marshal(map[string]string{
		"phone":    "09163503284",
		"password": "wrong-password",
	})
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)
	require.Equal(t, http.StatusUnauthorized, loginRec.Code)
}

func TestSignup_RejectsNonIranianPhone(t *testing.T) {
	router := newTestRouter(t)

	body, _ := json.Marshal(map[string]string{
		"first_name": "Ada",
		"last_name":  "Lovelace",
		"phone":      "+15550001111",
		"password":   "correct-horse-battery-staple",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), "Iranian mobile number")
}
