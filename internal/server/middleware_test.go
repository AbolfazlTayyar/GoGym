package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestLogger_SkipsOnlyHealthyProbes(t *testing.T) {
	tests := map[string]struct {
		path    string
		status  int
		wantLog bool
	}{
		"healthy probe":     {path: healthzPath, status: http.StatusOK, wantLog: false},
		"failing probe":     {path: healthzPath, status: http.StatusServiceUnavailable, wantLog: true},
		"other ok response": {path: APIV1Prefix + "/athletes", status: http.StatusOK, wantLog: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			var buf bytes.Buffer

			router := gin.New()
			router.Use(requestLogger(zerolog.New(&buf)))
			router.GET(tc.path, func(c *gin.Context) { c.Status(tc.status) })

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, http.NoBody))
			require.Equal(t, tc.status, rec.Code)

			if tc.wantLog {
				assert.Contains(t, buf.String(), `"path":"`+tc.path+`"`)
			} else {
				assert.Empty(t, buf.String())
			}
		})
	}
}
