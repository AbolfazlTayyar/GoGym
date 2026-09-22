package httpx_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AbolfazlTayyar/gogym/internal/httpx"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestContext returns a gin context writing into a fresh recorder, the
// way a handler would receive one.
func newTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	return c, rec
}

// decode unmarshals a response body into a generic map, so the assertions
// below see the literal JSON keys — including the ones that are supposed to
// be absent — rather than what a typed struct would paper over.
func decode(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	return body
}

type payload struct {
	Name string `json:"name"`
}

func TestOK(t *testing.T) {
	c, rec := newTestContext()

	httpx.OK(c, payload{Name: "Ada"})

	assert.Equal(t, http.StatusOK, rec.Code)

	body := decode(t, rec)
	assert.Equal(t, true, body["success"])
	assert.Equal(t, map[string]any{"name": "Ada"}, body["data"])
	assert.Nil(t, body["error"])
	assert.Contains(t, body, "error", "error is always present, even when null")
	assert.NotContains(t, body, "meta", "meta is omitted when empty")
}

func TestCreated(t *testing.T) {
	c, rec := newTestContext()

	httpx.Created(c, payload{Name: "Ada"})

	assert.Equal(t, http.StatusCreated, rec.Code)

	body := decode(t, rec)
	assert.Equal(t, true, body["success"])
	assert.Equal(t, map[string]any{"name": "Ada"}, body["data"])
	assert.Contains(t, body, "error")
	assert.NotContains(t, body, "meta")
}

func TestNoContent(t *testing.T) {
	c, rec := newTestContext()

	httpx.NoContent(c)

	// 200 with a null data, not a bodiless 204 — see the helper's comment.
	assert.Equal(t, http.StatusOK, rec.Code)

	body := decode(t, rec)
	assert.Equal(t, true, body["success"])
	assert.Contains(t, body, "data")
	assert.Nil(t, body["data"])
	assert.Nil(t, body["error"])
	assert.NotContains(t, body, "meta")
}

func TestOKWithMeta(t *testing.T) {
	c, rec := newTestContext()

	httpx.OKWithMeta(c, []payload{{Name: "Ada"}}, map[string]any{"total": 1})

	assert.Equal(t, http.StatusOK, rec.Code)

	body := decode(t, rec)
	assert.Equal(t, true, body["success"])
	assert.Equal(t, []any{map[string]any{"name": "Ada"}}, body["data"])
	assert.Equal(t, map[string]any{"total": float64(1)}, body["meta"])
}

func TestError(t *testing.T) {
	c, rec := newTestContext()

	httpx.Error(c, http.StatusConflict, httpx.CodeConflict, "phone already registered")

	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.True(t, c.IsAborted(), "an error response must abort, so middleware can use the same helper")

	body := decode(t, rec)
	assert.Equal(t, false, body["success"])
	assert.Contains(t, body, "data")
	assert.Nil(t, body["data"])
	assert.NotContains(t, body, "meta")

	errBody, ok := body["error"].(map[string]any)
	require.True(t, ok, "error must be an object")
	assert.Equal(t, httpx.CodeConflict, errBody["code"])
	assert.Equal(t, "phone already registered", errBody["message"])
	assert.NotContains(t, errBody, "fields", "fields is omitted when the helper carries none")
}

func TestErrorFields(t *testing.T) {
	c, rec := newTestContext()

	httpx.ErrorFields(c, http.StatusBadRequest, httpx.CodeValidationFailed, httpx.MsgInvalidRequest,
		map[string]string{"phone": "must be an Iranian mobile number"})

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	body := decode(t, rec)
	assert.Equal(t, false, body["success"])
	assert.Nil(t, body["data"])

	errBody, ok := body["error"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, httpx.CodeValidationFailed, errBody["code"])
	assert.Equal(t, httpx.MsgInvalidRequest, errBody["message"])
	assert.Equal(t, map[string]any{"phone": "must be an Iranian mobile number"}, errBody["fields"])
}

// TestErrorFields_EmptyMapOmitted covers calling ErrorFields with whatever
// ValidationFields returned, including nothing — the body should then be
// indistinguishable from an Error one.
func TestErrorFields_EmptyMapOmitted(t *testing.T) {
	for name, fields := range map[string]map[string]string{
		"nil":   nil,
		"empty": {},
	} {
		t.Run(name, func(t *testing.T) {
			c, rec := newTestContext()

			httpx.ErrorFields(c, http.StatusBadRequest, httpx.CodeValidationFailed, httpx.MsgInvalidRequest, fields)

			errBody, ok := decode(t, rec)["error"].(map[string]any)
			require.True(t, ok)
			assert.NotContains(t, errBody, "fields")
		})
	}
}
