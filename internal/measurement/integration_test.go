package measurement_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AbolfazlTayyar/gogym/internal/athlete"
	"github.com/AbolfazlTayyar/gogym/internal/coach"
	"github.com/AbolfazlTayyar/gogym/internal/httpx"
	"github.com/AbolfazlTayyar/gogym/internal/measurement"
	"github.com/AbolfazlTayyar/gogym/internal/testutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// newTestRouter wires coach and athlete too: measurements need a real coach's token and a real athlete.
func newTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()

	db := testutil.NewDB(t)

	coachRepo := coach.NewRepository(db)
	coachSvc := coach.NewService(coachRepo, "test-secret", time.Hour)
	athleteSvc := athlete.NewService(athlete.NewRepository(db))
	measurementSvc := measurement.NewService(measurement.NewRepository(db), athleteSvc)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	v1 := router.Group("/api/v1")
	protected := v1.Group("")
	protected.Use(coach.AuthMiddleware(coachSvc))

	coach.RegisterRoutes(v1, protected, coach.NewHandler(coachSvc, coachRepo))
	athlete.RegisterRoutes(protected, athlete.NewHandler(athleteSvc))
	measurement.RegisterRoutes(protected, measurement.NewHandler(measurementSvc))

	return router, db
}

type measurementResponse struct {
	ID        string   `json:"id"`
	AthleteID string   `json:"athlete_id"`
	Date      string   `json:"date"`
	Weight    *float64 `json:"weight"`
	Chest     *float64 `json:"chest"`
	Waist     *float64 `json:"waist"`
	Arm       *float64 `json:"arm"`
	Thigh     *float64 `json:"thigh"`
	Hip       *float64 `json:"hip"`
}

type envelope struct {
	Success bool             `json:"success"`
	Data    json.RawMessage  `json:"data"`
	Error   *httpx.ErrorBody `json:"error"`
}

func decodeEnvelope(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int) envelope {
	t.Helper()

	require.Equal(t, wantStatus, rec.Code, "body: %s", rec.Body.String())

	var env envelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))

	return env
}

func do(t *testing.T, router *gin.Engine, method, target, token, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, target, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec
}

func signupCoach(t *testing.T, router *gin.Engine, phone string) string {
	t.Helper()

	credentials := fmt.Sprintf(`{"first_name":"Test","last_name":"Coach","phone":%q,"password":"correct-horse-battery-staple"}`, phone)

	signup := do(t, router, http.MethodPost, "/api/v1/auth/signup", "", credentials)
	require.Equal(t, http.StatusCreated, signup.Code, "body: %s", signup.Body.String())

	env := decodeEnvelope(t, do(t, router, http.MethodPost, "/api/v1/auth/login", "", credentials), http.StatusOK)

	var payload struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.Unmarshal(env.Data, &payload))
	require.NotEmpty(t, payload.Token)

	return payload.Token
}

func createAthlete(t *testing.T, router *gin.Engine, token, phone string) string {
	t.Helper()

	body := fmt.Sprintf(`{"first_name":"Sara","last_name":"Ahmadi","phone":%q}`, phone)
	env := decodeEnvelope(t, do(t, router, http.MethodPost, "/api/v1/athletes", token, body), http.StatusCreated)

	var created struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(env.Data, &created))
	require.NotEmpty(t, created.ID)

	return created.ID
}

func measurementsPath(athleteID string) string {
	return "/api/v1/athletes/" + athleteID + "/measurements"
}

func addMeasurement(t *testing.T, router *gin.Engine, token, athleteID, body string) measurementResponse {
	t.Helper()

	env := decodeEnvelope(t, do(t, router, http.MethodPost, measurementsPath(athleteID), token, body), http.StatusCreated)
	assert.True(t, env.Success)
	assert.Nil(t, env.Error)

	var created measurementResponse
	require.NoError(t, json.Unmarshal(env.Data, &created))

	return created
}

func listMeasurements(t *testing.T, router *gin.Engine, token, athleteID string) []measurementResponse {
	t.Helper()

	env := decodeEnvelope(t, do(t, router, http.MethodGet, measurementsPath(athleteID), token, ""), http.StatusOK)
	assert.True(t, env.Success)

	var out []measurementResponse
	require.NoError(t, json.Unmarshal(env.Data, &out))

	return out
}

func dates(ms []measurementResponse) []string {
	out := make([]string, 0, len(ms))
	for _, m := range ms {
		out = append(out, m.Date)
	}
	return out
}

func TestMeasurements_ListedInDateOrder(t *testing.T) {
	router, _ := newTestRouter(t)
	token := signupCoach(t, router, "09121110101")
	athleteID := createAthlete(t, router, token, "09121230101")

	// Posted out of order, so insertion order and date order disagree.
	created := addMeasurement(t, router, token, athleteID, `{"date":"2026-09-20","weight":71.5,"waist":80}`)
	addMeasurement(t, router, token, athleteID, `{"date":"2026-08-01","weight":74,"waist":83.5,"chest":99}`)
	addMeasurement(t, router, token, athleteID, `{"date":"2026-08-25","weight":72.8}`)

	assert.Equal(t, athleteID, created.AthleteID)
	assert.Equal(t, "2026-09-20", created.Date, "the date round-trips as a plain date, not a timestamp")

	got := listMeasurements(t, router, token, athleteID)

	require.Equal(t, []string{"2026-08-01", "2026-08-25", "2026-09-20"}, dates(got))

	first := got[0]
	require.NotNil(t, first.Weight)
	assert.InDelta(t, 74, *first.Weight, 0)
	require.NotNil(t, first.Chest)
	assert.InDelta(t, 99, *first.Chest, 0)
	assert.Nil(t, first.Arm, "an unrecorded metric is null, so a chart skips it rather than plotting 0")

	last := got[2]
	require.NotNil(t, last.Waist)
	assert.InDelta(t, 80, *last.Waist, 0, "decimal values survive the numeric column")
}

func TestMeasurements_EmptyHistoryIsAnEmptyArray(t *testing.T) {
	router, _ := newTestRouter(t)
	token := signupCoach(t, router, "09121110102")
	athleteID := createAthlete(t, router, token, "09121230102")

	env := decodeEnvelope(t, do(t, router, http.MethodGet, measurementsPath(athleteID), token, ""), http.StatusOK)

	assert.JSONEq(t, `[]`, string(env.Data), "a chart gets [] to iterate, not null")
}

func TestMeasurements_AnotherCoachsAthleteIsNotFound(t *testing.T) {
	router, db := newTestRouter(t)

	owner := signupCoach(t, router, "09121110103")
	intruder := signupCoach(t, router, "09121110104")
	athleteID := createAthlete(t, router, owner, "09121230103")

	addMeasurement(t, router, owner, athleteID, `{"date":"2026-09-01","weight":70}`)

	t.Run("post is 404, not 403", func(t *testing.T) {
		rec := do(t, router, http.MethodPost, measurementsPath(athleteID), intruder, `{"date":"2026-09-02","weight":99}`)

		env := decodeEnvelope(t, rec, http.StatusNotFound)
		assert.False(t, env.Success)
		require.NotNil(t, env.Error)
		assert.Equal(t, httpx.CodeNotFound, env.Error.Code)

		var stored int64
		require.NoError(t, db.Model(&measurement.Measurement{}).Where("athlete_id = ?", athleteID).Count(&stored).Error)
		assert.Equal(t, int64(1), stored, "the rejected post wrote nothing")
	})

	t.Run("an invalid body still 404s, so validation can't probe for the athlete", func(t *testing.T) {
		rec := do(t, router, http.MethodPost, measurementsPath(athleteID), intruder, `{}`)
		decodeEnvelope(t, rec, http.StatusNotFound)
	})

	t.Run("list is 404", func(t *testing.T) {
		rec := do(t, router, http.MethodGet, measurementsPath(athleteID), intruder, "")
		env := decodeEnvelope(t, rec, http.StatusNotFound)
		assert.Equal(t, "null", string(env.Data), "the owner's measurements don't leak")
	})

	t.Run("an id that doesn't exist reads the same", func(t *testing.T) {
		rec := do(t, router, http.MethodGet, measurementsPath(uuid.NewString()), owner, "")
		decodeEnvelope(t, rec, http.StatusNotFound)
	})
}

func TestMeasurements_SoftDeletedAthleteIsNotFound(t *testing.T) {
	router, db := newTestRouter(t)
	token := signupCoach(t, router, "09121110105")
	athleteID := createAthlete(t, router, token, "09121230105")

	require.NoError(t, db.Delete(&athlete.Athlete{}, "id = ?", athleteID).Error)

	decodeEnvelope(t, do(t, router, http.MethodGet, measurementsPath(athleteID), token, ""), http.StatusNotFound)
	decodeEnvelope(t, do(t, router, http.MethodPost, measurementsPath(athleteID), token, `{"date":"2026-09-01","weight":70}`), http.StatusNotFound)
}

func TestMeasurements_RejectsBadInputWithFieldErrors(t *testing.T) {
	router, _ := newTestRouter(t)
	token := signupCoach(t, router, "09121110106")
	athleteID := createAthlete(t, router, token, "09121230106")

	rec := do(t, router, http.MethodPost, measurementsPath(athleteID), token, `{"date":"24/09/2026","waist":0.8}`)

	env := decodeEnvelope(t, rec, http.StatusBadRequest)
	require.NotNil(t, env.Error)
	assert.Equal(t, httpx.CodeValidationFailed, env.Error.Code)
	assert.Contains(t, env.Error.Fields, "date")
	assert.Contains(t, env.Error.Fields, "waist")

	assert.Empty(t, listMeasurements(t, router, token, athleteID), "a rejected post writes nothing")
}

func TestMeasurements_RequireAToken(t *testing.T) {
	router, _ := newTestRouter(t)

	rec := do(t, router, http.MethodGet, measurementsPath(uuid.NewString()), "", "")
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}
