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
