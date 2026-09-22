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

// These cover the handler paths that reject a request before it reaches the
// repository, so they need no database and run under `make test-unit`. The
// paths that do touch the repository are in integration_test.go.

// newHandlerTestRouter builds the real routes against a repository with a nil
// *gorm.DB. Any test here that reached the database would panic, which is the
// point: it would mean the request got further than the test assumes.
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

// requireErrorEnvelope asserts rec is a failure envelope with the expected
// status and code, and returns its error body.
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

// TestSignup_ReportsFieldLevelBindingErrors covers the envelope's reason for
// carrying a fields map: the client is told which field it got wrong, keyed
// by the json name it sent, not just that the request was bad.
func TestSignup_ReportsFieldLevelBindingErrors(t *testing.T) {
	rec := postJSONBody(newHandlerTestRouter(), "/api/v1/auth/signup",
		`{"first_name":"Ada","phone":"09372144430","password":"short"}`)

	errBody := requireErrorEnvelope(t, rec, http.StatusBadRequest, httpx.CodeValidationFailed)
	assert.Equal(t, map[string]string{
		"last_name": "is required",
		"password":  "must be at least 8 characters",
	}, errBody.Fields)
}

// TestLogin_MalformedBodyHasNoFields is the other binding failure: the body
// never parsed, so there is nothing to report per field and the key is absent.
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
