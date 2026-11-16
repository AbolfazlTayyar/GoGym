package membership_test

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
	"github.com/AbolfazlTayyar/gogym/internal/membership"
	"github.com/AbolfazlTayyar/gogym/internal/movement"
	"github.com/AbolfazlTayyar/gogym/internal/plan"
	"github.com/AbolfazlTayyar/gogym/internal/platform/httpx"
	"github.com/AbolfazlTayyar/gogym/internal/platform/testutil"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()

	db := testutil.NewDB(t)

	coachSvc := coach.NewService(coach.NewRepository(db), "test-secret", time.Hour)
	athleteSvc := athlete.NewService(athlete.NewRepository(db))
	planSvc := plan.NewService(plan.NewRepository(db), athleteSvc, movement.NewService(movement.NewRepository(db)), coachSvc)
	membershipSvc := membership.NewService(membership.NewRepository(db), planSvc)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	v1 := router.Group("/api/v1")
	protected := v1.Group("")
	protected.Use(coach.AuthMiddleware(coachSvc))

	coach.RegisterRoutes(v1, protected, coach.NewHandler(coachSvc, false))
	athlete.RegisterRoutes(protected, athlete.NewHandler(athleteSvc))
	plan.RegisterRoutes(v1, protected, plan.NewHandler(planSvc))
	membership.RegisterRoutes(protected, membership.NewHandler(membershipSvc))

	return router, db
}

type envelope struct {
	Success bool             `json:"success"`
	Data    json.RawMessage  `json:"data"`
	Error   *httpx.ErrorBody `json:"error"`
}

type attendanceResponse struct {
	Tracked  bool     `json:"tracked"`
	Sessions int      `json:"sessions"`
	Dates    []string `json:"dates"`
}

type currentResponse struct {
	PlanID    string  `json:"plan_id"`
	AthleteID string  `json:"athlete_id"`
	Visits    int     `json:"visits"`
	Sessions  int     `json:"sessions"`
	LastDate  *string `json:"last_date"`
}

func do(t *testing.T, router *gin.Engine, method, target, token, body string) envelopeRecorder {
	t.Helper()

	req := httptest.NewRequest(method, target, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return envelopeRecorder{t: t, rec: rec}
}

type envelopeRecorder struct {
	t   *testing.T
	rec *httptest.ResponseRecorder
}

func (r envelopeRecorder) expect(wantStatus int) envelope {
	r.t.Helper()

	require.Equal(r.t, wantStatus, r.rec.Code, "body: %s", r.rec.Body.String())

	var env envelope
	require.NoError(r.t, json.Unmarshal(r.rec.Body.Bytes(), &env))

	return env
}

func (r envelopeRecorder) attendance(wantStatus int) attendanceResponse {
	r.t.Helper()

	var a attendanceResponse
	require.NoError(r.t, json.Unmarshal(r.expect(wantStatus).Data, &a))

	return a
}

// createCoach goes through coach.Service because the API has no signup route.
func createCoach(t *testing.T, router *gin.Engine, db *gorm.DB, phone string) string {
	t.Helper()

	const password = "correct-horse-battery-staple"

	_, err := coach.NewService(coach.NewRepository(db), "", 0).Create(context.Background(), coach.CreateInput{
		FirstName: "Test",
		LastName:  "Coach",
		Phone:     phone,
		Password:  password,
	})
	require.NoError(t, err)

	credentials := fmt.Sprintf(`{"phone":%q,"password":%q}`, phone, password)
	r := do(t, router, http.MethodPost, "/api/v1/auth/login", "", credentials)
	r.expect(http.StatusOK)

	return testutil.Cookie(t, r.rec, coach.AuthCookieName)
}

func createID(t *testing.T, router *gin.Engine, token, path, body string) string {
	t.Helper()

	var created struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(do(t, router, http.MethodPost, path, token, body).expect(http.StatusCreated).Data, &created))

	return created.ID
}

func createPlan(t *testing.T, router *gin.Engine, token, athleteID, startDate string) string {
	t.Helper()

	body := fmt.Sprintf(`{"athlete_id":%q,"title":"Plan","start_date":%q}`, athleteID, startDate)
	return createID(t, router, token, "/api/v1/plans", body)
}

func attendancePath(planID string) string {
	return "/api/v1/plans/" + planID + "/attendance"
}

func visit(date string) string {
	return fmt.Sprintf(`{"date":%q}`, date)
}

func daysAgo(n int) string {
	return time.Now().UTC().AddDate(0, 0, -n).Format(time.DateOnly)
}

func TestAttendance_TrackRecordAndCount(t *testing.T) {
	router, db := newTestRouter(t)
	token := createCoach(t, router, db, "09121110201")
	athleteID := createID(t, router, token, "/api/v1/athletes", `{"first_name":"Sara","last_name":"Ahmadi","phone":"09121230201"}`)
	planID := createPlan(t, router, token, athleteID, daysAgo(30))
	path := attendancePath(planID)

	t.Run("a new plan is untracked and absent from current", func(t *testing.T) {
		got := do(t, router, http.MethodGet, path, token, "").attendance(http.StatusOK)
		assert.Equal(t, attendanceResponse{Tracked: false, Sessions: membership.SessionsPerPlan, Dates: []string{}}, got)

		env := do(t, router, http.MethodGet, "/api/v1/attendance/current", token, "").expect(http.StatusOK)
		assert.JSONEq(t, `[]`, string(env.Data))
	})

	t.Run("an untracked plan takes no visits", func(t *testing.T) {
		do(t, router, http.MethodPost, path+"/visits", token, visit(daysAgo(1))).expect(http.StatusNotFound)
	})

	do(t, router, http.MethodPut, path, token, "").attendance(http.StatusOK)

	t.Run("visits list oldest first", func(t *testing.T) {
		do(t, router, http.MethodPost, path+"/visits", token, visit(daysAgo(2))).attendance(http.StatusCreated)
		got := do(t, router, http.MethodPost, path+"/visits", token, visit(daysAgo(4))).attendance(http.StatusCreated)

		assert.True(t, got.Tracked)
		assert.Equal(t, []string{daysAgo(4), daysAgo(2)}, got.Dates)
	})

	t.Run("tracking again keeps the visits", func(t *testing.T) {
		got := do(t, router, http.MethodPut, path, token, "").attendance(http.StatusOK)
		assert.Len(t, got.Dates, 2)
	})

	t.Run("one visit per day", func(t *testing.T) {
		env := do(t, router, http.MethodPost, path+"/visits", token, visit(daysAgo(2))).expect(http.StatusBadRequest)
		assert.Equal(t, "is already recorded", env.Error.Fields["date"])
	})

	t.Run("no visit before the plan starts", func(t *testing.T) {
		env := do(t, router, http.MethodPost, path+"/visits", token, visit(daysAgo(31))).expect(http.StatusBadRequest)
		assert.Contains(t, env.Error.Fields["date"], "must be between")
	})

	t.Run("current shows the count and the last visit", func(t *testing.T) {
		var got []currentResponse
		require.NoError(t, json.Unmarshal(do(t, router, http.MethodGet, "/api/v1/attendance/current", token, "").expect(http.StatusOK).Data, &got))

		require.Len(t, got, 1)
		assert.Equal(t, planID, got[0].PlanID)
		assert.Equal(t, athleteID, got[0].AthleteID)
		assert.Equal(t, 2, got[0].Visits)
		require.NotNil(t, got[0].LastDate)
		assert.Equal(t, daysAgo(2), *got[0].LastDate)
	})

	t.Run("deleting a visit by date", func(t *testing.T) {
		got := do(t, router, http.MethodDelete, path+"/visits/"+daysAgo(4), token, "").attendance(http.StatusOK)
		assert.Equal(t, []string{daysAgo(2)}, got.Dates)

		do(t, router, http.MethodDelete, path+"/visits/"+daysAgo(4), token, "").expect(http.StatusNotFound)
		do(t, router, http.MethodDelete, path+"/visits/not-a-date", token, "").expect(http.StatusNotFound)
	})

	t.Run("a thirteenth visit is refused", func(t *testing.T) {
		for day := 5; day < 5+membership.SessionsPerPlan-1; day++ {
			do(t, router, http.MethodPost, path+"/visits", token, visit(daysAgo(day))).expect(http.StatusCreated)
		}

		env := do(t, router, http.MethodPost, path+"/visits", token, visit(daysAgo(0))).expect(http.StatusBadRequest)
		assert.Equal(t, "all 12 sessions are recorded", env.Error.Fields["date"])
	})

	t.Run("untracking deletes the visits", func(t *testing.T) {
		got := do(t, router, http.MethodDelete, path, token, "").attendance(http.StatusOK)
		assert.False(t, got.Tracked)

		var stored int64
		require.NoError(t, db.Model(&membership.Visit{}).Where("plan_id = ?", planID).Count(&stored).Error)
		assert.Zero(t, stored)
	})
}

func TestAttendance_CurrentFollowsTheCurrentPlan(t *testing.T) {
	router, db := newTestRouter(t)
	token := createCoach(t, router, db, "09121110202")
	athleteID := createID(t, router, token, "/api/v1/athletes", `{"first_name":"Sara","last_name":"Ahmadi","phone":"09121230202"}`)

	old := createPlan(t, router, token, athleteID, daysAgo(40))
	do(t, router, http.MethodPut, attendancePath(old), token, "").expect(http.StatusOK)

	upcoming := createPlan(t, router, token, athleteID, time.Now().UTC().AddDate(0, 0, 5).Format(time.DateOnly))
	do(t, router, http.MethodPut, attendancePath(upcoming), token, "").expect(http.StatusOK)

	var got []currentResponse
	require.NoError(t, json.Unmarshal(do(t, router, http.MethodGet, "/api/v1/attendance/current", token, "").expect(http.StatusOK).Data, &got))

	require.Len(t, got, 1, "a plan that hasn't started yet doesn't replace the current one")
	assert.Equal(t, old, got[0].PlanID)
	assert.Nil(t, got[0].LastDate)
}

func TestAttendance_AnotherCoachsPlanIsNotFound(t *testing.T) {
	router, db := newTestRouter(t)
	owner := createCoach(t, router, db, "09121110203")
	intruder := createCoach(t, router, db, "09121110204")
	athleteID := createID(t, router, owner, "/api/v1/athletes", `{"first_name":"Sara","last_name":"Ahmadi","phone":"09121230203"}`)
	planID := createPlan(t, router, owner, athleteID, daysAgo(10))
	path := attendancePath(planID)

	do(t, router, http.MethodPut, path, owner, "").expect(http.StatusOK)
	do(t, router, http.MethodPost, path+"/visits", owner, visit(daysAgo(1))).expect(http.StatusCreated)

	do(t, router, http.MethodGet, path, intruder, "").expect(http.StatusNotFound)
	do(t, router, http.MethodPut, path, intruder, "").expect(http.StatusNotFound)
	do(t, router, http.MethodDelete, path, intruder, "").expect(http.StatusNotFound)
	do(t, router, http.MethodPost, path+"/visits", intruder, visit(daysAgo(2))).expect(http.StatusNotFound)
	do(t, router, http.MethodDelete, path+"/visits/"+daysAgo(1), intruder, "").expect(http.StatusNotFound)

	env := do(t, router, http.MethodGet, "/api/v1/attendance/current", intruder, "").expect(http.StatusOK)
	assert.JSONEq(t, `[]`, string(env.Data))

	got := do(t, router, http.MethodGet, path, owner, "").attendance(http.StatusOK)
	assert.Equal(t, []string{daysAgo(1)}, got.Dates, "the intruder changed nothing")
}
