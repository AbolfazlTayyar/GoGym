package coach

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AbolfazlTayyar/gogym/internal/httpx"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newHandlerTestRouter uses a nil *gorm.DB on purpose: a test that reaches the database panics.
func newHandlerTestRouter() *gin.Engine {
	repo := NewRepository(nil)
	svc := NewService(repo, "test-secret", time.Hour)
	handler := NewHandler(svc, repo)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	v1 := router.Group("/api/v1")
	protected := v1.Group("")
	protected.Use(AuthMiddleware(svc))

	RegisterRoutes(v1, protected, handler)

	return router
}

func postJSONBody(router *gin.Engine, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func requireErrorEnvelope(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, wantCode string) *httpx.ErrorBody {
	t.Helper()

	require.Equal(t, wantStatus, rec.Code, "body: %s", rec.Body.String())

	var env struct {
		Success bool             `json:"success"`
		Data    json.RawMessage  `json:"data"`
		Error   *httpx.ErrorBody `json:"error"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))

	assert.False(t, env.Success)
	assert.Equal(t, "null", string(env.Data), "data is null on a failure envelope")
	assert.NotContains(t, rec.Body.String(), `"meta"`, "meta is omitted when empty")
	require.NotNil(t, env.Error)
	assert.Equal(t, wantCode, env.Error.Code)
	assert.NotEmpty(t, env.Error.Message)

	return env.Error
}

func TestSignup_ReportsFieldLevelBindingErrors(t *testing.T) {
	rec := postJSONBody(newHandlerTestRouter(), "/api/v1/auth/signup",
		`{"first_name":"Ada","phone":"09372144430","password":"short"}`)

	errBody := requireErrorEnvelope(t, rec, http.StatusBadRequest, httpx.CodeValidationFailed)
	assert.Equal(t, map[string]string{
		"last_name": "is required",
		"password":  "must be at least 8 characters",
	}, errBody.Fields)
}

func TestLogin_MalformedBodyHasNoFields(t *testing.T) {
	rec := postJSONBody(newHandlerTestRouter(), "/api/v1/auth/login", `{"phone":`)

	errBody := requireErrorEnvelope(t, rec, http.StatusBadRequest, httpx.CodeValidationFailed)
	assert.Empty(t, errBody.Fields)
	assert.NotContains(t, rec.Body.String(), `"fields"`)
}

func TestAuthMiddleware_Envelope(t *testing.T) {
	router := newHandlerTestRouter()

	for name, header := range map[string]string{
		"missing":   "",
		"malformed": "Token abc",
		"invalid":   "Bearer not-a-jwt",
	} {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/coaches/me", nil)
			if header != "" {
				req.Header.Set("Authorization", header)
			}
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			requireErrorEnvelope(t, rec, http.StatusUnauthorized, httpx.CodeUnauthorized)
		})
	}
}
