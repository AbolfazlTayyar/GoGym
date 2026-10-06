package plan_test

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
	"github.com/AbolfazlTayyar/gogym/internal/plan"
	"github.com/AbolfazlTayyar/gogym/internal/testutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// newTestRouter wires coach and athlete too: plans need a real coach's token and a real athlete.
func newTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()

	db := testutil.NewDB(t)

	coachRepo := coach.NewRepository(db)
	coachSvc := coach.NewService(coachRepo, "test-secret", time.Hour)
	athleteSvc := athlete.NewService(athlete.NewRepository(db))
	planSvc := plan.NewService(plan.NewRepository(db), athleteSvc)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	v1 := router.Group("/api/v1")
	protected := v1.Group("")
	protected.Use(coach.AuthMiddleware(coachSvc))

	coach.RegisterRoutes(v1, protected, coach.NewHandler(coachSvc, coachRepo))
	athlete.RegisterRoutes(protected, athlete.NewHandler(athleteSvc))
	plan.RegisterRoutes(protected, plan.NewHandler(planSvc))

	return router, db
}

type planResponse struct {
	ID        string `json:"id"`
	AthleteID string `json:"athlete_id"`
	Title     string `json:"title"`
	StartDate string `json:"start_date"`
	IsCurrent bool   `json:"is_current"`
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

// daysFromToday stays well clear of the today boundary, which the unit test covers with a fixed clock.
func daysFromToday(days int) time.Time {
	d := time.Now().AddDate(0, 0, days)
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
}

// insertPlan writes straight to the table: there is no create-plan endpoint yet.
func insertPlan(t *testing.T, db *gorm.DB, athleteID, title string, start time.Time) {
	t.Helper()

	require.NoError(t, db.Create(&plan.Plan{
		ID:        uuid.New(),
		AthleteID: uuid.MustParse(athleteID),
		Title:     title,
		StartDate: start,
	}).Error)
}

func plansPath(athleteID string) string {
	return "/api/v1/athletes/" + athleteID + "/plans"
}

func listPlans(t *testing.T, router *gin.Engine, token, athleteID string) []planResponse {
	t.Helper()

	env := decodeEnvelope(t, do(t, router, http.MethodGet, plansPath(athleteID), token, ""), http.StatusOK)
	assert.True(t, env.Success)
	assert.Nil(t, env.Error)

	var out []planResponse
	require.NoError(t, json.Unmarshal(env.Data, &out))

	return out
}

func TestPlans_ListedNewestFirstWithOneCurrent(t *testing.T) {
	router, db := newTestRouter(t)
	token := signupCoach(t, router, "09121110201")
	athleteID := createAthlete(t, router, token, "09121230201")

	// Inserted out of order, so insertion order and start_date order disagree.
	insertPlan(t, db, athleteID, "Cut phase 1", daysFromToday(-10))
	insertPlan(t, db, athleteID, "Cut phase 2", daysFromToday(20))
	insertPlan(t, db, athleteID, "Base", daysFromToday(-90))

	got := listPlans(t, router, token, athleteID)
	require.Len(t, got, 3)

	titles := []string{got[0].Title, got[1].Title, got[2].Title}
	assert.Equal(t, []string{"Cut phase 2", "Cut phase 1", "Base"}, titles, "newest start_date first")
	assert.Equal(t, daysFromToday(-10).Format(time.DateOnly), got[1].StartDate, "start_date is a plain date, not a timestamp")
	assert.Equal(t, athleteID, got[1].AthleteID)

	var current []string
	for _, p := range got {
		if p.IsCurrent {
			current = append(current, p.Title)
		}
	}
	assert.Equal(t, []string{"Cut phase 1"}, current,
		"exactly one current plan: the latest that has started, not the future-dated one")
}

func TestPlans_OnlyFuturePlansMeansNoneCurrent(t *testing.T) {
	router, db := newTestRouter(t)
	token := signupCoach(t, router, "09121110202")
	athleteID := createAthlete(t, router, token, "09121230202")

	insertPlan(t, db, athleteID, "Next block", daysFromToday(14))

	got := listPlans(t, router, token, athleteID)
	require.Len(t, got, 1)
	assert.False(t, got[0].IsCurrent)
}

func TestPlans_EmptyListIsAnEmptyArray(t *testing.T) {
	router, _ := newTestRouter(t)
	token := signupCoach(t, router, "09121110203")
	athleteID := createAthlete(t, router, token, "09121230203")

	env := decodeEnvelope(t, do(t, router, http.MethodGet, plansPath(athleteID), token, ""), http.StatusOK)

	assert.JSONEq(t, `[]`, string(env.Data))
}

func TestPlans_AnotherCoachsAthleteIsNotFound(t *testing.T) {
	router, db := newTestRouter(t)

	owner := signupCoach(t, router, "09121110204")
	intruder := signupCoach(t, router, "09121110205")
	athleteID := createAthlete(t, router, owner, "09121230204")

	insertPlan(t, db, athleteID, "Base", daysFromToday(-30))

	t.Run("list is 404, not 403", func(t *testing.T) {
		env := decodeEnvelope(t, do(t, router, http.MethodGet, plansPath(athleteID), intruder, ""), http.StatusNotFound)
		assert.False(t, env.Success)
		require.NotNil(t, env.Error)
		assert.Equal(t, httpx.CodeNotFound, env.Error.Code)
		assert.Equal(t, "null", string(env.Data), "the owner's plans don't leak")
	})

	t.Run("an id that doesn't exist reads the same", func(t *testing.T) {
		decodeEnvelope(t, do(t, router, http.MethodGet, plansPath(uuid.NewString()), owner, ""), http.StatusNotFound)
	})

	t.Run("a soft-deleted athlete reads the same", func(t *testing.T) {
		require.NoError(t, db.Delete(&athlete.Athlete{}, "id = ?", athleteID).Error)
		decodeEnvelope(t, do(t, router, http.MethodGet, plansPath(athleteID), owner, ""), http.StatusNotFound)
	})
}

func TestPlans_RequireAToken(t *testing.T) {
	router, _ := newTestRouter(t)

	rec := do(t, router, http.MethodGet, plansPath(uuid.NewString()), "", "")
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}
