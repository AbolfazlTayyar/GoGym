package measurement_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AbolfazlTayyar/gogym/internal/athlete"
	"github.com/AbolfazlTayyar/gogym/internal/coach"
	"github.com/AbolfazlTayyar/gogym/internal/measurement"
	"github.com/AbolfazlTayyar/gogym/internal/platform/httpx"
	"github.com/AbolfazlTayyar/gogym/internal/platform/testutil"
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

	coachSvc := coach.NewService(coach.NewRepository(db), "test-secret", time.Hour)
	athleteSvc := athlete.NewService(athlete.NewRepository(db))
	measurementSvc := measurement.NewService(measurement.NewRepository(db), athleteSvc)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	v1 := router.Group("/api/v1")
	protected := v1.Group("")
	protected.Use(coach.AuthMiddleware(coachSvc))

	coach.RegisterRoutes(v1, protected, coach.NewHandler(coachSvc))
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

// createCoach goes through coach.Service because the API has no signup route.
func createCoach(t *testing.T, router *gin.Engine, db *gorm.DB, phone string) string {
	t.Helper()

	const password = "correct-horse-battery-staple"

	// Create never issues a token, so the service needs no JWT settings.
	_, err := coach.NewService(coach.NewRepository(db), "", 0).Create(context.Background(), coach.CreateInput{
		FirstName: "Test",
		LastName:  "Coach",
		Phone:     phone,
		Password:  password,
	})
	require.NoError(t, err)

	credentials := fmt.Sprintf(`{"phone":%q,"password":%q}`, phone, password)
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

func measurementByIDPath(id string) string {
	return "/api/v1/measurements/" + id
}

func findMeasurement(t *testing.T, db *gorm.DB, id string) measurement.Measurement {
	t.Helper()

	var m measurement.Measurement
	require.NoError(t, db.First(&m, "id = ?", id).Error)

	return m
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
	router, db := newTestRouter(t)
	token := createCoach(t, router, db, "09121110101")
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
	router, db := newTestRouter(t)
	token := createCoach(t, router, db, "09121110102")
	athleteID := createAthlete(t, router, token, "09121230102")

	env := decodeEnvelope(t, do(t, router, http.MethodGet, measurementsPath(athleteID), token, ""), http.StatusOK)

	assert.JSONEq(t, `[]`, string(env.Data), "a chart gets [] to iterate, not null")
}

func TestMeasurements_AnotherCoachsAthleteIsNotFound(t *testing.T) {
	router, db := newTestRouter(t)

	owner := createCoach(t, router, db, "09121110103")
	intruder := createCoach(t, router, db, "09121110104")
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
	token := createCoach(t, router, db, "09121110105")
	athleteID := createAthlete(t, router, token, "09121230105")
	created := addMeasurement(t, router, token, athleteID, `{"date":"2026-09-01","weight":70}`)

	require.NoError(t, db.Delete(&athlete.Athlete{}, "id = ?", athleteID).Error)

	decodeEnvelope(t, do(t, router, http.MethodGet, measurementsPath(athleteID), token, ""), http.StatusNotFound)
	decodeEnvelope(t, do(t, router, http.MethodPost, measurementsPath(athleteID), token, `{"date":"2026-09-01","weight":70}`), http.StatusNotFound)
	decodeEnvelope(t, do(t, router, http.MethodPut, measurementByIDPath(created.ID), token, `{"date":"2026-09-01","weight":71}`), http.StatusNotFound)
	decodeEnvelope(t, do(t, router, http.MethodDelete, measurementByIDPath(created.ID), token, ""), http.StatusNotFound)
}

func TestMeasurements_UpdateAndDelete(t *testing.T) {
	router, db := newTestRouter(t)
	token := createCoach(t, router, db, "09121110107")
	athleteID := createAthlete(t, router, token, "09121230107")

	created := addMeasurement(t, router, token, athleteID, `{"date":"2026-09-20","weight":71.5,"waist":80}`)
	kept := addMeasurement(t, router, token, athleteID, `{"date":"2026-08-01","weight":74}`)
	path := measurementByIDPath(created.ID)

	t.Run("update replaces the date and every metric", func(t *testing.T) {
		env := decodeEnvelope(t, do(t, router, http.MethodPut, path, token, `{"date":"2026-07-15","weight":70.9,"hip":98}`), http.StatusOK)

		var updated measurementResponse
		require.NoError(t, json.Unmarshal(env.Data, &updated))
		assert.Equal(t, created.ID, updated.ID)
		assert.Equal(t, athleteID, updated.AthleteID)
		assert.Equal(t, "2026-07-15", updated.Date)
		require.NotNil(t, updated.Weight)
		assert.InDelta(t, 70.9, *updated.Weight, 0)
		require.NotNil(t, updated.Hip)
		assert.InDelta(t, 98, *updated.Hip, 0)
		assert.Nil(t, updated.Waist, "a metric left out is cleared")

		stored := findMeasurement(t, db, created.ID)
		assert.Nil(t, stored.Waist)
		assert.True(t, stored.UpdatedAt.After(stored.CreatedAt), "updated_at moves on an edit")

		assert.Equal(t, []string{"2026-07-15", "2026-08-01"}, dates(listMeasurements(t, router, token, athleteID)),
			"the history re-sorts by the new date")
	})

	t.Run("update validates like create", func(t *testing.T) {
		env := decodeEnvelope(t, do(t, router, http.MethodPut, path, token, `{"date":"2026-07-15"}`), http.StatusBadRequest)
		require.NotNil(t, env.Error)
		assert.Equal(t, httpx.CodeValidationFailed, env.Error.Code)
		assert.Contains(t, env.Error.Fields, "weight")

		stored := findMeasurement(t, db, created.ID)
		require.NotNil(t, stored.Weight)
		assert.InDelta(t, 70.9, *stored.Weight, 0, "a rejected update writes nothing")
	})

	t.Run("delete removes the row for good", func(t *testing.T) {
		env := decodeEnvelope(t, do(t, router, http.MethodDelete, path, token, ""), http.StatusOK)
		assert.True(t, env.Success)
		assert.Equal(t, "null", string(env.Data))

		var rows int64
		require.NoError(t, db.Model(&measurement.Measurement{}).Where("id = ?", created.ID).Count(&rows).Error)
		assert.Zero(t, rows, "a hard delete leaves no row behind")

		assert.Equal(t, []string{kept.Date}, dates(listMeasurements(t, router, token, athleteID)))
	})

	t.Run("a deleted measurement is gone", func(t *testing.T) {
		decodeEnvelope(t, do(t, router, http.MethodPut, path, token, `{"date":"2026-07-15","weight":70}`), http.StatusNotFound)
		decodeEnvelope(t, do(t, router, http.MethodDelete, path, token, ""), http.StatusNotFound)
	})
}

// Another coach's measurement gets the same 404 as one that doesn't exist, whatever the body holds.
func TestMeasurements_OnlyTheOwnerMayChangeAMeasurement(t *testing.T) {
	router, db := newTestRouter(t)

	owner := createCoach(t, router, db, "09121110108")
	intruder := createCoach(t, router, db, "09121110109")
	athleteID := createAthlete(t, router, owner, "09121230108")

	created := addMeasurement(t, router, owner, athleteID, `{"date":"2026-09-01","weight":70}`)
	path := measurementByIDPath(created.ID)

	t.Run("update and delete are 404, not 403", func(t *testing.T) {
		for _, rec := range []*httptest.ResponseRecorder{
			do(t, router, http.MethodPut, path, intruder, `{"date":"2026-09-02","weight":99}`),
			do(t, router, http.MethodPut, path, intruder, `{}`),
			do(t, router, http.MethodDelete, path, intruder, ""),
		} {
			env := decodeEnvelope(t, rec, http.StatusNotFound)
			require.NotNil(t, env.Error)
			assert.Equal(t, httpx.CodeNotFound, env.Error.Code)
		}

		stored := findMeasurement(t, db, created.ID)
		require.NotNil(t, stored.Weight)
		assert.InDelta(t, 70, *stored.Weight, 0, "neither changed nor deleted")
	})

	t.Run("an id that doesn't exist reads the same", func(t *testing.T) {
		unknown := measurementByIDPath(uuid.NewString())
		decodeEnvelope(t, do(t, router, http.MethodPut, unknown, owner, `{"date":"2026-09-02","weight":71}`), http.StatusNotFound)
		decodeEnvelope(t, do(t, router, http.MethodDelete, unknown, owner, ""), http.StatusNotFound)
	})
}

func TestMeasurements_RejectsBadInputWithFieldErrors(t *testing.T) {
	router, db := newTestRouter(t)
	token := createCoach(t, router, db, "09121110106")
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
