package server

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestHealthz_UnreachableDBHidesDetailFromCaller(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Port 1 refuses at once, so the ping fails fast with a driver error naming the host.
	const unreachableHost = "127.0.0.1"
	gormDB, err := gorm.Open(
		postgres.Open("postgres://u:p@"+unreachableHost+":1/db?sslmode=disable&connect_timeout=1"),
		&gorm.Config{DisableAutomaticPing: true},
	)
	require.NoError(t, err)

	var logs bytes.Buffer
	router := gin.New()
	router.Use(requestLogger(zerolog.New(&logs)))
	router.GET(healthzPath, healthzHandler(gormDB))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, healthzPath, http.NoBody))

	require.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.JSONEq(t, `{"status":"unavailable","reason":"database unreachable"}`, rec.Body.String())
	assert.NotContains(t, rec.Body.String(), unreachableHost)

	assert.Contains(t, logs.String(), `"level":"error"`)
	assert.Contains(t, logs.String(), unreachableHost, "the real error should reach the logs")
}

func TestHealthz_ReusesPingResultWithinTTL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	pingErr := errors.New("connection refused")
	pings := 0
	now := time.Now()
	checker := &healthChecker{
		ping: func(context.Context) error { pings++; return pingErr },
		ttl:  healthzCacheTTL,
		now:  func() time.Time { return now },
	}

	var logs bytes.Buffer
	router := gin.New()
	router.Use(requestLogger(zerolog.New(&logs)))
	router.GET(healthzPath, newHealthzHandler(checker))

	get := func() int {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, healthzPath, http.NoBody))
		return rec.Code
	}

	require.Equal(t, http.StatusServiceUnavailable, get())
	logs.Reset()
	require.Equal(t, http.StatusServiceUnavailable, get())
	assert.Equal(t, 1, pings, "a second request within the TTL should not ping again")
	assert.Contains(t, logs.String(), pingErr.Error(), "a cached failure should still log its cause")

	pingErr = nil
	now = now.Add(healthzCacheTTL)
	require.Equal(t, http.StatusOK, get())
	assert.Equal(t, 2, pings, "an expired result should be re-checked")
}
