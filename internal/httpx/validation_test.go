package httpx_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AbolfazlTayyar/gogym/internal/httpx"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type bindRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	Password  string `json:"password" binding:"required,min=8"`
}

// bindErr runs body through gin's binder the way a handler does, and returns
// the error it produced.
func bindErr(t *testing.T, body string) error {
	t.Helper()

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(body)))
	c.Request.Header.Set("Content-Type", "application/json")

	var req bindRequest
	err := c.ShouldBindJSON(&req)
	require.Error(t, err, "the fixture body is supposed to fail binding")
	return err
}

// TestValidationFields_UsesJSONNames is the reason httpx registers a tag-name
// func on gin's validator: without it these keys would be "FirstName" and
// "Password", which is not what the client sent.
func TestValidationFields_UsesJSONNames(t *testing.T) {
	fields := httpx.ValidationFields(bindErr(t, `{"password":"short"}`))

	assert.Equal(t, map[string]string{
		"first_name": "is required",
		"password":   "must be at least 8 characters",
	}, fields)
}

func TestValidationFields_NonFieldErrorReturnsNil(t *testing.T) {
	// Malformed JSON fails before per-field validation runs, so there is
	// no field detail to report.
	assert.Nil(t, httpx.ValidationFields(bindErr(t, `{"first_name":`)))
}

// TestValidationFields_IntoErrorFields is the end-to-end shape a handler
// produces from a binding failure.
func TestValidationFields_IntoErrorFields(t *testing.T) {
	err := bindErr(t, `{"first_name":"Ada"}`)

	c, rec := newTestContext()
	httpx.ErrorFields(c, http.StatusBadRequest, httpx.CodeValidationFailed, httpx.MsgInvalidRequest, httpx.ValidationFields(err))

	require.Equal(t, http.StatusBadRequest, rec.Code)

	var body struct {
		Success bool             `json:"success"`
		Data    *json.RawMessage `json:"data"`
		Error   *httpx.ErrorBody `json:"error"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

	assert.False(t, body.Success)
	require.NotNil(t, body.Error)
	assert.Equal(t, httpx.CodeValidationFailed, body.Error.Code)
	assert.Equal(t, map[string]string{"password": "is required"}, body.Error.Fields)
}
