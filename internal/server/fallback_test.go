package server

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

// newFallbackTestRouter wires the three responses Gin produces without a
// handler the same way New does, minus the database-backed routes.
func newFallbackTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(gin.CustomRecovery(recoveryHandler))
	router.HandleMethodNotAllowed = true
	router.NoRoute(notFoundHandler)
	router.NoMethod(methodNotAllowedHandler)

	router.GET(APIV1Prefix+"/panics", func(*gin.Context) { panic("boom") })

	return router
}

func requireErrorEnvelope(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, wantCode string) {
	t.Helper()

	require.Equal(t, wantStatus, rec.Code, "body: %s", rec.Body.String())

	var env struct {
		Success bool             `json:"success"`
		Data    json.RawMessage  `json:"data"`
		Error   *httpx.ErrorBody `json:"error"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env), "body was not an envelope: %s", rec.Body.String())

	assert.False(t, env.Success)
	assert.Equal(t, "null", string(env.Data))
	require.NotNil(t, env.Error)
	assert.Equal(t, wantCode, env.Error.Code)
	assert.NotEmpty(t, env.Error.Message)
}

func TestFallbacks(t *testing.T) {
	tests := map[string]struct {
		method     string
		path       string
		wantStatus int
		wantCode   string
	}{
		"unmatched path": {
			method:     http.MethodGet,
			path:       APIV1Prefix + "/does-not-exist",
			wantStatus: http.StatusNotFound,
			wantCode:   httpx.CodeNotFound,
		},
		"wrong method": {
			method:     http.MethodDelete,
			path:       APIV1Prefix + "/panics",
			wantStatus: http.StatusMethodNotAllowed,
			wantCode:   httpx.CodeMethodNotAllowed,
		},
		"panic": {
			method:     http.MethodGet,
			path:       APIV1Prefix + "/panics",
			wantStatus: http.StatusInternalServerError,
			wantCode:   httpx.CodeInternalError,
		},
	}

	router := newFallbackTestRouter()

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, nil))

			requireErrorEnvelope(t, rec, tc.wantStatus, tc.wantCode)
		})
	}
}
