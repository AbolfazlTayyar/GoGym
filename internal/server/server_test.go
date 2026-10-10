package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AbolfazlTayyar/gogym/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_ClientIPHonorsOnlyTrustedProxies(t *testing.T) {
	const (
		proxyIP   = "10.0.0.5"
		callerIP  = "203.0.113.9"
		spoofedIP = "1.2.3.4"
	)

	tests := map[string]struct {
		trustedProxies []string
		remoteIP       string
		header         string
		want           string
	}{
		"forwarded-for ignored when no proxy is trusted": {
			remoteIP: callerIP,
			header:   "X-Forwarded-For",
			want:     callerIP,
		},
		"real-ip ignored when no proxy is trusted": {
			remoteIP: callerIP,
			header:   "X-Real-IP",
			want:     callerIP,
		},
		"forwarded-for ignored from an untrusted peer": {
			trustedProxies: []string{proxyIP},
			remoteIP:       callerIP,
			header:         "X-Forwarded-For",
			want:           callerIP,
		},
		"forwarded-for honored from a trusted proxy": {
			trustedProxies: []string{proxyIP},
			remoteIP:       proxyIP,
			header:         "X-Forwarded-For",
			want:           spoofedIP,
		},
		"forwarded-for honored from a trusted CIDR": {
			trustedProxies: []string{"10.0.0.0/8"},
			remoteIP:       proxyIP,
			header:         "X-Forwarded-For",
			want:           spoofedIP,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			// Construction never touches the database, so no connection is needed.
			router, err := New(config.Config{TrustedProxies: tc.trustedProxies}, nil, zerolog.Nop())
			require.NoError(t, err)
			router.GET("/test/client-ip", func(c *gin.Context) { c.String(http.StatusOK, c.ClientIP()) })

			req := httptest.NewRequest(http.MethodGet, "/test/client-ip", http.NoBody)
			req.RemoteAddr = tc.remoteIP + ":5555"
			req.Header.Set(tc.header, spoofedIP)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, tc.want, rec.Body.String())
		})
	}
}

func TestNew_RejectsInvalidTrustedProxy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	_, err := New(config.Config{TrustedProxies: []string{"not-an-ip"}}, nil, zerolog.Nop())

	assert.Error(t, err)
}

func TestNew_RecoveredPanicIsLogged(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var logs bytes.Buffer

	router, err := New(config.Config{}, nil, zerolog.New(&logs))
	require.NoError(t, err)
	router.GET("/test/panics", func(*gin.Context) { panic("boom") })

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/test/panics", http.NoBody))

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.NotContains(t, rec.Body.String(), "boom")
	assert.Contains(t, logs.String(), `"level":"error"`)
	assert.Contains(t, logs.String(), `"errors":["panic: boom"]`)
}
