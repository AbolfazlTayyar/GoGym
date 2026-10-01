package athlete

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AbolfazlTayyar/gogym/internal/httpx"
	"github.com/AbolfazlTayyar/gogym/internal/tenant"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The repository holds a nil *gorm.DB on purpose: a test that reaches a query panics.

// newTestContext leaves the request unauthenticated when coachID is uuid.Nil.
func newTestContext(t *testing.T, method, target string, body string, coachID uuid.UUID) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(method, target, bytes.NewReader([]byte(body)))
	c.Request.Header.Set("Content-Type", "application/json")

	if coachID != uuid.Nil {
		tenant.SetCoachID(c, coachID)
	}

	return c, rec
}

func newTestHandler() *Handler {
	return NewHandler(NewService(NewRepository(nil)))
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
	require.NotNil(t, env.Error)
	assert.Equal(t, wantCode, env.Error.Code)
	assert.NotEmpty(t, env.Error.Message)

	return env.Error
}

func TestCreate_ReportsFieldLevelErrors(t *testing.T) {
	c, rec := newTestContext(t, http.MethodPost, "/api/v1/athletes",
		`{"first_name":"Sara","phone":"12345"}`, uuid.New())

	newTestHandler().Create(c)

	errBody := requireErrorEnvelope(t, rec, http.StatusBadRequest, httpx.CodeValidationFailed)
	assert.Equal(t, msgRequired, errBody.Fields["last_name"])
	assert.Contains(t, errBody.Fields["phone"], "Iranian mobile number")
}

func TestCreate_MalformedBodyHasNoFields(t *testing.T) {
	c, rec := newTestContext(t, http.MethodPost, "/api/v1/athletes", `{"first_name":`, uuid.New())

	newTestHandler().Create(c)

	errBody := requireErrorEnvelope(t, rec, http.StatusBadRequest, httpx.CodeValidationFailed)
	assert.Empty(t, errBody.Fields)
	assert.NotContains(t, rec.Body.String(), `"fields"`)
}

func TestList_RejectsBadQueryParams(t *testing.T) {
	for name, tc := range map[string]struct {
		target    string
		wantField string
	}{
		"unknown athlete type": {target: "/api/v1/athletes?athlete_type=everyone", wantField: "athlete_type"},
		"non-numeric limit":    {target: "/api/v1/athletes?limit=lots", wantField: "limit"},
		"non-numeric offset":   {target: "/api/v1/athletes?offset=next", wantField: "offset"},
	} {
		t.Run(name, func(t *testing.T) {
			c, rec := newTestContext(t, http.MethodGet, tc.target, "", uuid.New())

			newTestHandler().List(c)

			errBody := requireErrorEnvelope(t, rec, http.StatusBadRequest, httpx.CodeValidationFailed)
			assert.Contains(t, errBody.Fields, tc.wantField)
		})
	}
}

// TestUnauthenticated covers the handlers' own guard, independent of the auth middleware.
func TestUnauthenticated(t *testing.T) {
	h := newTestHandler()

	t.Run("create", func(t *testing.T) {
		c, rec := newTestContext(t, http.MethodPost, "/api/v1/athletes", `{}`, uuid.Nil)
		h.Create(c)
		requireErrorEnvelope(t, rec, http.StatusUnauthorized, httpx.CodeUnauthorized)
	})

	t.Run("list", func(t *testing.T) {
		c, rec := newTestContext(t, http.MethodGet, "/api/v1/athletes", "", uuid.Nil)
		h.List(c)
		requireErrorEnvelope(t, rec, http.StatusUnauthorized, httpx.CodeUnauthorized)
	})
}
