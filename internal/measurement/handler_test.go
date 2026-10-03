package measurement

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AbolfazlTayyar/gogym/internal/athlete"
	"github.com/AbolfazlTayyar/gogym/internal/httpx"
	"github.com/AbolfazlTayyar/gogym/internal/tenant"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The repositories hold a nil *gorm.DB on purpose: a test that reaches a query panics.

// newTestContext leaves the request unauthenticated when coachID is uuid.Nil.
func newTestContext(t *testing.T, method, athleteID, body string, coachID uuid.UUID) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(method, "/api/v1/athletes/"+athleteID+"/measurements", bytes.NewReader([]byte(body)))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: paramAthleteID, Value: athleteID}}

	if coachID != uuid.Nil {
		tenant.SetCoachID(c, coachID)
	}

	return c, rec
}

func newTestHandler() *Handler {
	athleteSvc := athlete.NewService(athlete.NewRepository(nil))
	return NewHandler(NewService(NewRepository(nil), athleteSvc))
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

	return env.Error
}

func TestCreate_MalformedBodyHasNoFields(t *testing.T) {
	c, rec := newTestContext(t, http.MethodPost, uuid.NewString(), `{"date":`, uuid.New())

	newTestHandler().Create(c)

	errBody := requireErrorEnvelope(t, rec, http.StatusBadRequest, httpx.CodeValidationFailed)
	assert.Empty(t, errBody.Fields)
}

func TestMalformedAthleteIDIsNotFound(t *testing.T) {
	h := newTestHandler()

	t.Run("create", func(t *testing.T) {
		c, rec := newTestContext(t, http.MethodPost, "not-a-uuid", `{"date":"2026-09-24","weight":70}`, uuid.New())
		h.Create(c)
		requireErrorEnvelope(t, rec, http.StatusNotFound, httpx.CodeNotFound)
	})

	t.Run("list", func(t *testing.T) {
		c, rec := newTestContext(t, http.MethodGet, "not-a-uuid", "", uuid.New())
		h.List(c)
		requireErrorEnvelope(t, rec, http.StatusNotFound, httpx.CodeNotFound)
	})
}

// TestUnauthenticated covers the handlers' own guard, independent of the auth middleware.
func TestUnauthenticated(t *testing.T) {
	h := newTestHandler()

	t.Run("create", func(t *testing.T) {
		c, rec := newTestContext(t, http.MethodPost, uuid.NewString(), `{}`, uuid.Nil)
		h.Create(c)
		requireErrorEnvelope(t, rec, http.StatusUnauthorized, httpx.CodeUnauthorized)
	})

	t.Run("list", func(t *testing.T) {
		c, rec := newTestContext(t, http.MethodGet, uuid.NewString(), "", uuid.Nil)
		h.List(c)
		requireErrorEnvelope(t, rec, http.StatusUnauthorized, httpx.CodeUnauthorized)
	})
}
