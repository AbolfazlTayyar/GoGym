package plan_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AbolfazlTayyar/gogym/internal/athlete"
	"github.com/AbolfazlTayyar/gogym/internal/coach"
	"github.com/AbolfazlTayyar/gogym/internal/movement"
	"github.com/AbolfazlTayyar/gogym/internal/plan"
	"github.com/AbolfazlTayyar/gogym/internal/platform/httpx"
	"github.com/AbolfazlTayyar/gogym/internal/platform/testutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

// newTestRouter wires coach and athlete too: plans need a real coach's token and a real athlete.
func newTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()

	db := testutil.NewDB(t)

	coachSvc := coach.NewService(coach.NewRepository(db), "test-secret", time.Hour)
	athleteSvc := athlete.NewService(athlete.NewRepository(db))
	planSvc := plan.NewService(plan.NewRepository(db), athleteSvc, movement.NewService(movement.NewRepository(db)))

	gin.SetMode(gin.TestMode)
	router := gin.New()

	v1 := router.Group("/api/v1")
	protected := v1.Group("")
	protected.Use(coach.AuthMiddleware(coachSvc))

	coach.RegisterRoutes(v1, protected, coach.NewHandler(coachSvc, false))
	athlete.RegisterRoutes(protected, athlete.NewHandler(athleteSvc))
	plan.RegisterRoutes(protected, plan.NewHandler(planSvc))

	return router, db
}

type planResponse struct {
	ID         string   `json:"id"`
	AthleteID  string   `json:"athlete_id"`
	Title      string   `json:"title"`
	StartDate  string   `json:"start_date"`
	IsCurrent  bool     `json:"is_current"`
	DayLabels  []string `json:"day_labels"`
	BlockCount int      `json:"block_count"`
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
	rec := do(t, router, http.MethodPost, "/api/v1/auth/login", "", credentials)
	decodeEnvelope(t, rec, http.StatusOK)

	// The tests send the cookie's token back as a Bearer header, which the middleware also accepts.
	return testutil.Cookie(t, rec, coach.AuthCookieName)
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

// insertPlan writes straight to the table, so read-side tests don't depend on the builder's validation.
func insertPlan(t *testing.T, db *gorm.DB, athleteID, title string, start time.Time) uuid.UUID {
	t.Helper()

	id := uuid.New()
	require.NoError(t, db.Create(&plan.Plan{
		ID:        id,
		AthleteID: uuid.MustParse(athleteID),
		Title:     title,
		StartDate: start,
	}).Error)

	return id
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
	token := createCoach(t, router, db, "09121110201")
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
	token := createCoach(t, router, db, "09121110202")
	athleteID := createAthlete(t, router, token, "09121230202")

	insertPlan(t, db, athleteID, "Next block", daysFromToday(14))

	got := listPlans(t, router, token, athleteID)
	require.Len(t, got, 1)
	assert.False(t, got[0].IsCurrent)
}

func TestPlans_EmptyListIsAnEmptyArray(t *testing.T) {
	router, db := newTestRouter(t)
	token := createCoach(t, router, db, "09121110203")
	athleteID := createAthlete(t, router, token, "09121230203")

	env := decodeEnvelope(t, do(t, router, http.MethodGet, plansPath(athleteID), token, ""), http.StatusOK)

	assert.JSONEq(t, `[]`, string(env.Data))
}

func TestPlans_AnotherCoachsAthleteIsNotFound(t *testing.T) {
	router, db := newTestRouter(t)

	owner := createCoach(t, router, db, "09121110204")
	intruder := createCoach(t, router, db, "09121110205")
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

func TestPlans_SummariesCarryDayLabelsAndBlockCount(t *testing.T) {
	router, db := newTestRouter(t)
	token := createCoach(t, router, db, "09121110206")
	athleteID := createAthlete(t, router, token, "09121230206")
	squat := insertMovement(t, db, "Back squat", "strength")

	full := insertPlan(t, db, athleteID, "Full", daysFromToday(-10))
	insertPlan(t, db, athleteID, "Empty", daysFromToday(10))

	// Days inserted out of order, holding 2, 3 and 0 blocks; one block is a two-movement superset.
	_, err := insertDay(db, full, "C", 2)
	require.NoError(t, err)
	dayA, err := insertDay(db, full, "A", 0)
	require.NoError(t, err)
	dayB, err := insertDay(db, full, "B", 1)
	require.NoError(t, err)
	for i := range 2 {
		insertBlock(t, db, dayA, i, 3, nil)
	}
	for i := range 3 {
		block := insertBlock(t, db, dayB, i, 3, nil)
		insertBlockMovement(t, db, block, squat, 12, 0)
		insertBlockMovement(t, db, block, squat, 10, 1)
	}

	env := decodeEnvelope(t, do(t, router, http.MethodGet, plansPath(athleteID), token, ""), http.StatusOK)
	var got []planResponse
	require.NoError(t, json.Unmarshal(env.Data, &got))
	require.Len(t, got, 2)

	assert.Equal(t, "Empty", got[0].Title)
	assert.Equal(t, []string{}, got[0].DayLabels)
	assert.Equal(t, 0, got[0].BlockCount)
	assert.Contains(t, string(env.Data), `"day_labels":[]`, "an empty plan's labels are [], never null")

	assert.Equal(t, "Full", got[1].Title)
	assert.Equal(t, []string{"A", "B", "C"}, got[1].DayLabels, "in order_index order")
	assert.Equal(t, 5, got[1].BlockCount, "a superset counts once")
}

func TestPlans_SummariesTakeTheSameQueriesForOneOrManyPlans(t *testing.T) {
	router, db := newTestRouter(t)
	token := createCoach(t, router, db, "09121110207")
	one := createAthlete(t, router, token, "09121230207")
	five := createAthlete(t, router, token, "09121230208")

	addPlan := func(athleteID string, n int) {
		planID := insertPlan(t, db, athleteID, fmt.Sprintf("Plan %d", n), daysFromToday(-30*n))
		day, err := insertDay(db, planID, "A", 0)
		require.NoError(t, err)
		insertBlock(t, db, day, 0, 3, nil)
	}
	addPlan(one, 1)
	for n := 1; n <= 5; n++ {
		addPlan(five, n)
	}

	rec := recordSQL(db)
	listPlans(t, router, token, one)
	forOne := len(rec.take())
	listPlans(t, router, token, five)
	forFive := len(rec.take())

	assert.Equal(t, forOne, forFive)
}

func currentPlans(t *testing.T, router *gin.Engine, token string) []planResponse {
	t.Helper()

	env := decodeEnvelope(t, do(t, router, http.MethodGet, "/api/v1/plans/current", token, ""), http.StatusOK)
	var out []planResponse
	require.NoError(t, json.Unmarshal(env.Data, &out))
	return out
}

func TestCurrentPlans_OnePerAthleteWithAStartedPlan(t *testing.T) {
	router, db := newTestRouter(t)
	token := createCoach(t, router, db, "09121110211")
	withTwo := createAthlete(t, router, token, "09121230211")
	withOne := createAthlete(t, router, token, "09121230212")
	createAthlete(t, router, token, "09121230213") // no plan at all
	onlyFuture := createAthlete(t, router, token, "09121230214")

	started := insertPlan(t, db, withTwo, "Mehr", daysFromToday(-10))
	insertPlan(t, db, withTwo, "Aban", daysFromToday(10))
	insertPlan(t, db, withTwo, "Shahrivar", daysFromToday(-40))
	other := insertPlan(t, db, withOne, "Mehr", daysFromToday(-5))
	insertPlan(t, db, onlyFuture, "Aban", daysFromToday(10))
	day, err := insertDay(db, started, "A", 0)
	require.NoError(t, err)
	insertBlock(t, db, day, 0, 3, nil)

	got := currentPlans(t, router, token)
	require.Len(t, got, 2)

	byAthlete := map[string]planResponse{}
	for _, p := range got {
		byAthlete[p.AthleteID] = p
		assert.True(t, p.IsCurrent)
	}
	assert.Equal(t, started.String(), byAthlete[withTwo].ID, "the started plan, not the future or older one")
	assert.Equal(t, []string{"A"}, byAthlete[withTwo].DayLabels)
	assert.Equal(t, 1, byAthlete[withTwo].BlockCount)
	assert.Equal(t, other.String(), byAthlete[withOne].ID)
}

func TestCurrentPlans_OnlyTheCoachsLiveAthletes(t *testing.T) {
	router, db := newTestRouter(t)
	coachA := createCoach(t, router, db, "09121110215")
	coachB := createCoach(t, router, db, "09121110216")
	mine := createAthlete(t, router, coachA, "09121230215")
	deleted := createAthlete(t, router, coachA, "09121230216")
	theirs := createAthlete(t, router, coachB, "09121230217")

	minePlan := insertPlan(t, db, mine, "Mehr", daysFromToday(-3))
	insertPlan(t, db, deleted, "Mehr", daysFromToday(-3))
	insertPlan(t, db, theirs, "Mehr", daysFromToday(-3))
	require.NoError(t, db.Delete(&athlete.Athlete{}, "id = ?", deleted).Error)

	got := currentPlans(t, router, coachA)
	require.Len(t, got, 1, "neither coach B's athlete nor a deleted athlete shows up")
	assert.Equal(t, minePlan.String(), got[0].ID)

	assert.Len(t, currentPlans(t, router, coachB), 1)
}

func TestCurrentPlans_TakeTheSameQueriesForOneOrManyAthletes(t *testing.T) {
	router, db := newTestRouter(t)
	one := createCoach(t, router, db, "09121110217")
	many := createCoach(t, router, db, "09121110218")
	insertPlan(t, db, createAthlete(t, router, one, "09121230218"), "Mehr", daysFromToday(-3))
	for i := range 4 {
		insertPlan(t, db, createAthlete(t, router, many, fmt.Sprintf("0912123022%d", i)), "Mehr", daysFromToday(-3))
	}

	rec := recordSQL(db)
	currentPlans(t, router, one)
	forOne := len(rec.take())
	currentPlans(t, router, many)
	forMany := len(rec.take())

	assert.Equal(t, forOne, forMany)
}

func TestCurrentPlans_EmptyIsAnEmptyArray(t *testing.T) {
	router, db := newTestRouter(t)
	token := createCoach(t, router, db, "09121110219")

	env := decodeEnvelope(t, do(t, router, http.MethodGet, "/api/v1/plans/current", token, ""), http.StatusOK)
	assert.JSONEq(t, `[]`, string(env.Data))
}

func TestUpdatePlan_RoundTripsAndClearsAnOmittedNote(t *testing.T) {
	router, db := newTestRouter(t)
	token := createCoach(t, router, db, "09121110501")
	athleteID := createAthlete(t, router, token, "09121230501")
	created := decodeEnvelope(t, do(t, router, http.MethodPost, "/api/v1/plans", token,
		fmt.Sprintf(`{"athlete_id":%q,"title":"Mehr","start_date":"2026-09-23","note":"deload week 4"}`, athleteID)), http.StatusCreated)
	var p struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(created.Data, &p))
	_, err := insertDay(db, uuid.MustParse(p.ID), "A", 0)
	require.NoError(t, err)

	env := decodeEnvelope(t, do(t, router, http.MethodPut, planPath(p.ID), token,
		`{"title":"برنامه آبان","start_date":"2026-10-23"}`), http.StatusOK)
	var updated struct {
		Title     string  `json:"title"`
		StartDate string  `json:"start_date"`
		Note      *string `json:"note"`
		Days      []any   `json:"days"`
	}
	require.NoError(t, json.Unmarshal(env.Data, &updated))
	assert.Equal(t, "برنامه آبان", updated.Title)
	assert.Equal(t, "2026-10-23", updated.StartDate)
	assert.Nil(t, updated.Note, "an omitted note is cleared")
	assert.Len(t, updated.Days, 1, "the response carries the plan's days")

	got, _ := getPlan(t, router, token, p.ID)
	assert.Equal(t, "برنامه آبان", got.Title)

	bad := decodeEnvelope(t, do(t, router, http.MethodPut, planPath(p.ID), token, `{"title":" ","start_date":"23/10/2026"}`), http.StatusBadRequest)
	require.NotNil(t, bad.Error)
	assert.Equal(t, httpx.MsgFieldRequired, bad.Error.Fields["title"])
	assert.Equal(t, httpx.MsgDateFormat, bad.Error.Fields["start_date"])
}

func TestUpdatePlan_MovingTheCurrentPlanAheadMakesThePreviousCurrent(t *testing.T) {
	router, db := newTestRouter(t)
	token := createCoach(t, router, db, "09121110502")
	athleteID := createAthlete(t, router, token, "09121230502")
	previous := insertPlan(t, db, athleteID, "Shahrivar", daysFromToday(-40))
	current := insertPlan(t, db, athleteID, "Mehr", daysFromToday(-10))

	body := fmt.Sprintf(`{"title":"Mehr","start_date":%q}`, daysFromToday(5).Format(time.DateOnly))
	decodeEnvelope(t, do(t, router, http.MethodPut, planPath(current.String()), token, body), http.StatusOK)

	for _, p := range listPlans(t, router, token, athleteID) {
		assert.Equal(t, p.ID == previous.String(), p.IsCurrent, p.Title)
	}
}

func TestDeletePlan_RemovesItsWholeTree(t *testing.T) {
	router, db := newTestRouter(t)
	token := createCoach(t, router, db, "09121110503")
	athleteID := createAthlete(t, router, token, "09121230503")
	planID := insertPlan(t, db, athleteID, "Mehr", daysFromToday(-10))
	squat := insertMovement(t, db, "Back squat", "strength")
	dayID, err := insertDay(db, planID, "A", 0)
	require.NoError(t, err)
	blockID := insertBlock(t, db, dayID, 0, 3, nil)
	insertBlockMovement(t, db, blockID, squat, 10, 0)

	env := decodeEnvelope(t, do(t, router, http.MethodDelete, planPath(planID.String()), token, ""), http.StatusOK)
	assert.Equal(t, "null", string(env.Data))

	decodeEnvelope(t, do(t, router, http.MethodGet, planPath(planID.String()), token, ""), http.StatusNotFound)
	count := func(table, column string, id uuid.UUID) int64 {
		var n int64
		require.NoError(t, db.Table(table).Where(column+" = ?", id).Count(&n).Error)
		return n
	}
	assert.Zero(t, count("day", "plan_id", planID))
	assert.Zero(t, count("block", "day_id", dayID))
	assert.Zero(t, count("block_movement", "block_id", blockID))
	assert.Equal(t, int64(1), count("movement", "id", squat), "the library movement stays")
}

func TestUpdateAndDeletePlan_AnotherCoachsPlanIsNotFound(t *testing.T) {
	router, db := newTestRouter(t)
	owner := createCoach(t, router, db, "09121110504")
	intruder := createCoach(t, router, db, "09121110505")
	planID := insertPlan(t, db, createAthlete(t, router, owner, "09121230504"), "Mehr", daysFromToday(-10))

	decodeEnvelope(t, do(t, router, http.MethodPut, planPath(planID.String()), intruder,
		`{"title":"x","start_date":"2026-10-23"}`), http.StatusNotFound)
	decodeEnvelope(t, do(t, router, http.MethodDelete, planPath(planID.String()), intruder, ""), http.StatusNotFound)

	got, _ := getPlan(t, router, owner, planID.String())
	assert.Equal(t, "Mehr", got.Title)
}

func saveDayPath(dayID uuid.UUID) string {
	return "/api/v1/days/" + dayID.String() + "/blocks"
}

// line builds one block of a save body: one movement per id, each with its reps.
func line(sets int, movements ...any) map[string]any {
	parts := make([]map[string]any, 0, len(movements)/2)
	for i := 0; i < len(movements); i += 2 {
		parts = append(parts, map[string]any{"movement_id": movements[i], "reps": movements[i+1]})
	}
	return map[string]any{"sets": sets, "movements": parts}
}

func saveBody(t *testing.T, lines ...map[string]any) string {
	t.Helper()

	if lines == nil {
		return "[]"
	}
	raw, err := json.Marshal(lines)
	require.NoError(t, err)
	return string(raw)
}

// dayLines reads a day back as one string per line, e.g. "3: lunge 12 + bridge 15".
func dayLines(t *testing.T, router *gin.Engine, token string, planID uuid.UUID, label string) []string {
	t.Helper()

	got, _ := getPlan(t, router, token, planID.String())
	for _, d := range got.Days {
		if d.Label != label {
			continue
		}
		out := make([]string, 0, len(d.Blocks))
		for _, b := range d.Blocks {
			parts := make([]string, 0, len(b.Movements))
			for _, m := range b.Movements {
				parts = append(parts, fmt.Sprintf("%s %d", m.Name, *m.Reps))
			}
			out = append(out, fmt.Sprintf("%d: %s", b.Sets, strings.Join(parts, " + ")))
		}
		return out
	}
	t.Fatalf("no day %s", label)
	return nil
}

type saveFixture struct {
	router                        *gin.Engine
	db                            *gorm.DB
	token                         string
	planID, dayID                 uuid.UUID
	lunge, bridge, squat, rowerID uuid.UUID
}

func newSaveFixture(t *testing.T, coachPhone, athletePhone string) saveFixture {
	t.Helper()

	router, db := newTestRouter(t)
	token := createCoach(t, router, db, coachPhone)
	planID := insertPlan(t, db, createAthlete(t, router, token, athletePhone), "Mehr", daysFromToday(-5))
	dayID, err := insertDay(db, planID, "C", 2)
	require.NoError(t, err)

	return saveFixture{
		router: router, db: db, token: token, planID: planID, dayID: dayID,
		lunge:   insertMovement(t, db, "lunge", "strength"),
		bridge:  insertMovement(t, db, "bridge", "strength"),
		squat:   insertMovement(t, db, "squat", "strength"),
		rowerID: insertMovement(t, db, "rower", "cardio"),
	}
}

func (f saveFixture) save(t *testing.T, body string, wantStatus int) envelope {
	t.Helper()
	return decodeEnvelope(t, do(t, f.router, http.MethodPut, saveDayPath(f.dayID), f.token, body), wantStatus)
}

func TestSaveDay_ReplacesTheBlocksInOrder(t *testing.T) {
	f := newSaveFixture(t, "09121110601", "09121230601")

	f.save(t, saveBody(t,
		line(3, f.lunge, 12), line(3, f.bridge, 15), line(4, f.squat, 8), line(1, f.rowerID, 20),
	), http.StatusOK)
	require.Equal(t, []string{"3: lunge 12", "3: bridge 15", "4: squat 8", "1: rower 20"},
		dayLines(t, f.router, f.token, f.planID, "C"))

	// Line 3 removed, 1 and 2 swapped, line 4 turned into a two-movement superset.
	env := f.save(t, saveBody(t,
		line(3, f.bridge, 15), line(3, f.lunge, 12), line(2, f.rowerID, 20, f.squat, 10),
	), http.StatusOK)

	assert.Equal(t, []string{"3: bridge 15", "3: lunge 12", "2: rower 20 + squat 10"},
		dayLines(t, f.router, f.token, f.planID, "C"))

	var day struct {
		Label  string `json:"label"`
		Blocks []struct {
			OrderIndex int `json:"order_index"`
			Movements  []struct {
				Name         string `json:"name"`
				OrderInBlock int    `json:"order_in_block"`
			} `json:"movements"`
		} `json:"blocks"`
	}
	require.NoError(t, json.Unmarshal(env.Data, &day))
	assert.Equal(t, "C", day.Label, "the response is the saved day")
	require.Len(t, day.Blocks, 3)
	assert.Equal(t, 2, day.Blocks[2].OrderIndex)
	assert.Equal(t, "squat", day.Blocks[2].Movements[1].Name)
	assert.Equal(t, 1, day.Blocks[2].Movements[1].OrderInBlock)
}

func TestSaveDay_AnEmptyArrayClearsTheDay(t *testing.T) {
	f := newSaveFixture(t, "09121110602", "09121230602")
	f.save(t, saveBody(t, line(3, f.lunge, 12)), http.StatusOK)

	f.save(t, "[]", http.StatusOK)

	assert.Empty(t, dayLines(t, f.router, f.token, f.planID, "C"))
}

func TestSaveDay_OneBadLineLeavesTheDayAsItWas(t *testing.T) {
	f := newSaveFixture(t, "09121110603", "09121230603")
	f.save(t, saveBody(t, line(3, f.lunge, 12), line(3, f.bridge, 15)), http.StatusOK)
	var before int64
	require.NoError(t, f.db.Table("block").Where("day_id = ?", f.dayID).Count(&before).Error)

	other, err := coach.NewService(coach.NewRepository(f.db), "", 0).Create(context.Background(), coach.CreateInput{
		FirstName: "Other", LastName: "Coach", Phone: "09121110604", Password: "correct-horse-battery-staple",
	})
	require.NoError(t, err)
	theirs := uuid.New()
	require.NoError(t, f.db.Create(&movement.Movement{ID: theirs, CoachID: &other.ID, Name: "theirs"}).Error)

	env := f.save(t, saveBody(t,
		line(3, f.squat, 8), line(3, f.lunge, 12), line(0, f.bridge, 15), line(3, f.rowerID, 20, theirs, 10),
	), http.StatusBadRequest)

	require.NotNil(t, env.Error)
	assert.Equal(t, map[string]string{
		"[2].sets":                     "must be between 1 and 50",
		"[3].movements[1].movement_id": "not found",
	}, env.Error.Fields, "every bad entry, by position, and another coach's movement reads as not found")

	var after int64
	require.NoError(t, f.db.Table("block").Where("day_id = ?", f.dayID).Count(&after).Error)
	assert.Equal(t, before, after)
	assert.Equal(t, []string{"3: lunge 12", "3: bridge 15"}, dayLines(t, f.router, f.token, f.planID, "C"))
}

func TestSaveDay_ShapeErrors(t *testing.T) {
	f := newSaveFixture(t, "09121110605", "09121230605")

	env := f.save(t, `[{"sets":3,"movements":[]},{"sets":3,"movements":[{"movement_id":"nope"}]}]`, http.StatusBadRequest)
	assert.Equal(t, map[string]string{
		"[0].movements":                "at least one movement is required",
		"[1].movements[0].movement_id": "must be a UUID",
	}, env.Error.Fields)

	tooMany := make([]map[string]any, 41)
	for i := range tooMany {
		tooMany[i] = line(3, f.lunge, 12)
	}
	env = f.save(t, saveBody(t, tooMany...), http.StatusBadRequest)
	assert.Equal(t, httpx.CodeValidationFailed, env.Error.Code)
}

func TestSaveDay_AMovementDeletedSinceStillSaves(t *testing.T) {
	f := newSaveFixture(t, "09121110606", "09121230606")
	f.save(t, saveBody(t, line(3, f.lunge, 12)), http.StatusOK)
	require.NoError(t, f.db.Delete(&movement.Movement{}, "id = ?", f.lunge).Error)

	f.save(t, saveBody(t, line(3, f.bridge, 15), line(4, f.lunge, 10)), http.StatusOK)
	assert.Equal(t, []string{"3: bridge 15", "4: lunge 10"}, dayLines(t, f.router, f.token, f.planID, "C"))

	t.Run("but a newly added deleted movement doesn't", func(t *testing.T) {
		require.NoError(t, f.db.Delete(&movement.Movement{}, "id = ?", f.squat).Error)
		env := f.save(t, saveBody(t, line(3, f.bridge, 15), line(3, f.squat, 8)), http.StatusBadRequest)
		assert.Equal(t, map[string]string{"[1].movements[0].movement_id": "not found"}, env.Error.Fields)
	})
}

func TestSaveDay_TakesTheSameQueriesForAShortOrLongDay(t *testing.T) {
	f := newSaveFixture(t, "09121110607", "09121230607")
	short := saveBody(t, line(3, f.lunge, 12), line(3, f.bridge, 15), line(3, f.squat, 8))
	long := make([]map[string]any, 9)
	for i := range long {
		long[i] = line(3, f.lunge, 12, f.bridge, 10)
	}

	rec := recordSQL(f.db)
	f.save(t, short, http.StatusOK)
	forShort := len(rec.take())
	f.save(t, saveBody(t, long...), http.StatusOK)
	forLong := len(rec.take())

	assert.Equal(t, forShort, forLong)
}

func TestSaveDay_ConcurrentSavesLeaveOneWholeDay(t *testing.T) {
	f := newSaveFixture(t, "09121110611", "09121230611")
	first := saveBody(t, line(3, f.lunge, 12), line(3, f.bridge, 15))
	second := saveBody(t, line(4, f.squat, 8))

	for range 10 {
		var wg sync.WaitGroup
		for _, body := range []string{first, second} {
			wg.Add(1)
			go func() {
				defer wg.Done()
				rec := do(t, f.router, http.MethodPut, saveDayPath(f.dayID), f.token, body)
				assert.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			}()
		}
		wg.Wait()

		got := dayLines(t, f.router, f.token, f.planID, "C")
		assert.Contains(t, [][]string{{"3: lunge 12", "3: bridge 15"}, {"4: squat 8"}}, got,
			"one save wins whole; the day never holds both")
	}
}

func TestDeleteDay_LeavesTheOtherSlotsAlone(t *testing.T) {
	f := newSaveFixture(t, "09121110608", "09121230608")
	_, err := insertDay(f.db, f.planID, "A", 0)
	require.NoError(t, err)
	_, err = insertDay(f.db, f.planID, "B", 1)
	require.NoError(t, err)
	f.save(t, saveBody(t, line(3, f.lunge, 12)), http.StatusOK)

	env := decodeEnvelope(t, do(t, f.router, http.MethodDelete, "/api/v1/days/"+f.dayID.String(), f.token, ""), http.StatusOK)
	assert.Equal(t, "null", string(env.Data))

	got, _ := getPlan(t, f.router, f.token, f.planID.String())
	require.Len(t, got.Days, 2)
	assert.Equal(t, []string{"A", "B"}, []string{got.Days[0].Label, got.Days[1].Label})
	assert.Equal(t, []int{0, 1}, []int{got.Days[0].OrderIndex, got.Days[1].OrderIndex})
	var blocks int64
	require.NoError(t, f.db.Table("block").Where("day_id = ?", f.dayID).Count(&blocks).Error)
	assert.Zero(t, blocks)
}

func TestSaveAndDeleteDay_AnotherCoachsDayIsNotFound(t *testing.T) {
	f := newSaveFixture(t, "09121110609", "09121230609")
	f.save(t, saveBody(t, line(3, f.lunge, 12)), http.StatusOK)
	intruder := createCoach(t, f.router, f.db, "09121110610")

	decodeEnvelope(t, do(t, f.router, http.MethodPut, saveDayPath(f.dayID), intruder, "[]"), http.StatusNotFound)
	decodeEnvelope(t, do(t, f.router, http.MethodDelete, "/api/v1/days/"+f.dayID.String(), intruder, ""), http.StatusNotFound)

	assert.Equal(t, []string{"3: lunge 12"}, dayLines(t, f.router, f.token, f.planID, "C"))
}

type copiedPlan struct {
	ID              string  `json:"id"`
	AthleteID       string  `json:"athlete_id"`
	Title           string  `json:"title"`
	StartDate       string  `json:"start_date"`
	SourcePlanID    *string `json:"source_plan_id"`
	SourcePlanTitle *string `json:"source_plan_title"`
}

func createPlan(t *testing.T, router *gin.Engine, token, body string, wantStatus int) (copiedPlan, envelope) {
	t.Helper()

	env := decodeEnvelope(t, do(t, router, http.MethodPost, "/api/v1/plans", token, body), wantStatus)
	var p copiedPlan
	if wantStatus == http.StatusCreated {
		require.NoError(t, json.Unmarshal(env.Data, &p))
	}
	return p, env
}

// treeOf is a plan's content without ids: what a copy has to match.
func treeOf(t *testing.T, router *gin.Engine, token, planID string) []string {
	t.Helper()

	got, _ := getPlan(t, router, token, planID)
	var out []string
	for _, d := range got.Days {
		for _, b := range d.Blocks {
			parts := make([]string, 0, len(b.Movements))
			for _, m := range b.Movements {
				load := ""
				if m.Load != nil {
					load = " " + *m.Load
				}
				parts = append(parts, fmt.Sprintf("%s:%d%s", m.Name, *m.Reps, load))
			}
			notes := ""
			if b.Notes != nil {
				notes = *b.Notes
			}
			out = append(out, fmt.Sprintf("%s/%d %d×%d rest %d %q [%s]",
				d.Label, d.OrderIndex, b.OrderIndex, b.Sets, *b.RestSeconds, notes, strings.Join(parts, " + ")))
		}
	}
	return out
}

func TestCopyPlan_MatchesTheSourceApartFromIdsTitleAndDate(t *testing.T) {
	f := newSaveFixture(t, "09121110701", "09121230701")
	dayA, err := insertDay(f.db, f.planID, "A", 0)
	require.NoError(t, err)
	superset := "superset"
	ss := insertBlock(t, f.db, dayA, 0, 3, &superset)
	insertBlockMovement(t, f.db, ss, f.lunge, 12, 0)
	insertBlockMovement(t, f.db, ss, f.bridge, 15, 1)
	single := insertBlock(t, f.db, f.dayID, 0, 4, nil)
	insertBlockMovement(t, f.db, single, f.squat, 8, 0)
	require.NoError(t, f.db.Delete(&movement.Movement{}, "id = ?", f.squat).Error)

	source, _ := getPlan(t, f.router, f.token, f.planID.String())
	body := fmt.Sprintf(`{"athlete_id":%q,"title":"برنامه آبان","start_date":"2026-10-23","source_plan_id":%q}`,
		source.AthleteID, f.planID)
	copied, env := createPlan(t, f.router, f.token, body, http.StatusCreated)

	assert.NotEqual(t, f.planID.String(), copied.ID)
	assert.Equal(t, "برنامه آبان", copied.Title)
	assert.Equal(t, "2026-10-23", copied.StartDate)
	require.NotNil(t, copied.SourcePlanID)
	assert.Equal(t, f.planID.String(), *copied.SourcePlanID, "the response is the copy's full detail")
	assert.Contains(t, string(env.Data), `"name":"squat"`, "a movement deleted from the library is copied as is")

	assert.Equal(t, treeOf(t, f.router, f.token, f.planID.String()), treeOf(t, f.router, f.token, copied.ID))

	got, _ := getPlan(t, f.router, f.token, copied.ID)
	for _, d := range got.Days {
		for _, sd := range source.Days {
			assert.NotEqual(t, sd.ID, d.ID, "every row is new")
		}
	}
}

func TestCopyPlan_ToAnotherAthleteAndEditingTheCopy(t *testing.T) {
	f := newSaveFixture(t, "09121110702", "09121230702")
	f.save(t, saveBody(t, line(3, f.lunge, 12), line(3, f.bridge, 15)), http.StatusOK)
	other := createAthlete(t, f.router, f.token, "09121230703")

	copied, _ := createPlan(t, f.router, f.token, fmt.Sprintf(
		`{"athlete_id":%q,"title":"Aban","start_date":"2026-10-23","source_plan_id":%q}`, other, f.planID), http.StatusCreated)
	assert.Equal(t, other, copied.AthleteID)

	got, _ := getPlan(t, f.router, f.token, copied.ID)
	require.Len(t, got.Days, 1)
	decodeEnvelope(t, do(t, f.router, http.MethodPut, "/api/v1/days/"+got.Days[0].ID+"/blocks", f.token,
		saveBody(t, line(5, f.squat, 5))), http.StatusOK)

	assert.Equal(t, []string{"3: lunge 12", "3: bridge 15"}, dayLines(t, f.router, f.token, f.planID, "C"),
		"the source is untouched")
	assert.Equal(t, []string{"5: squat 5"}, dayLines(t, f.router, f.token, uuid.MustParse(copied.ID), "C"))
}

func TestCopyPlan_NamesItsSourceUntilTheSourceIsDeleted(t *testing.T) {
	f := newSaveFixture(t, "09121110704", "09121230704")
	source, _ := getPlan(t, f.router, f.token, f.planID.String())
	copied, _ := createPlan(t, f.router, f.token, fmt.Sprintf(
		`{"athlete_id":%q,"title":"Aban","start_date":"2026-10-23","source_plan_id":%q}`, source.AthleteID, f.planID), http.StatusCreated)

	_, raw := getPlan(t, f.router, f.token, copied.ID)
	var detail copiedPlan
	require.NoError(t, json.Unmarshal(raw, &detail))
	require.NotNil(t, detail.SourcePlanTitle)
	assert.Equal(t, "Mehr", *detail.SourcePlanTitle)

	decodeEnvelope(t, do(t, f.router, http.MethodDelete, planPath(f.planID.String()), f.token, ""), http.StatusOK)

	_, raw = getPlan(t, f.router, f.token, copied.ID)
	require.NoError(t, json.Unmarshal(raw, &detail))
	assert.Nil(t, detail.SourcePlanID)
	assert.Nil(t, detail.SourcePlanTitle)
	assert.Contains(t, string(raw), `"source_plan_id":null`)
}

func TestCopyPlan_AnotherCoachsSourceIsAFieldErrorAndCreatesNothing(t *testing.T) {
	f := newSaveFixture(t, "09121110705", "09121230705")
	intruder := createCoach(t, f.router, f.db, "09121110706")
	theirAthlete := createAthlete(t, f.router, intruder, "09121230706")

	_, env := createPlan(t, f.router, intruder, fmt.Sprintf(
		`{"athlete_id":%q,"title":"x","start_date":"2026-10-23","source_plan_id":%q}`, theirAthlete, f.planID), http.StatusBadRequest)
	assert.Equal(t, map[string]string{"source_plan_id": "not found"}, env.Error.Fields)

	_, env = createPlan(t, f.router, intruder, fmt.Sprintf(
		`{"athlete_id":%q,"title":"x","start_date":"2026-10-23","source_plan_id":"nope"}`, theirAthlete), http.StatusBadRequest)
	assert.Equal(t, map[string]string{"source_plan_id": "must be a UUID"}, env.Error.Fields)

	assert.Empty(t, listPlans(t, f.router, intruder, theirAthlete))
}

func TestCreatePlan_WithoutASourceIsAnEmptyPlan(t *testing.T) {
	router, db := newTestRouter(t)
	token := createCoach(t, router, db, "09121110707")
	athleteID := createAthlete(t, router, token, "09121230707")

	created, env := createPlan(t, router, token, fmt.Sprintf(`{"athlete_id":%q,"title":"Mehr","start_date":"2026-09-23"}`, athleteID), http.StatusCreated)
	assert.Contains(t, string(env.Data), `"days":[]`)
	assert.Nil(t, created.SourcePlanID)
	assert.Nil(t, created.SourcePlanTitle)

	_, raw := getPlan(t, router, token, created.ID)
	assert.Contains(t, string(raw), `"source_plan_id":null`)
	assert.Contains(t, string(raw), `"source_plan_title":null`)
}

func TestCopyPlan_TakesTheSameQueriesForOneOrFiveDays(t *testing.T) {
	f := newSaveFixture(t, "09121110708", "09121230708")
	athleteID := createAthlete(t, f.router, f.token, "09121230709")
	bigPlan := insertPlan(t, f.db, athleteID, "Big", daysFromToday(-3))
	for i, label := range []string{"A", "B", "C", "D", "E"} {
		day, err := insertDay(f.db, bigPlan, label, i)
		require.NoError(t, err)
		for j := range 3 {
			block := insertBlock(t, f.db, day, j, 3, nil)
			insertBlockMovement(t, f.db, block, f.lunge, 12, 0)
		}
	}
	f.save(t, saveBody(t, line(3, f.lunge, 12)), http.StatusOK)

	copyOf := func(source uuid.UUID) int {
		rec := recordSQL(f.db)
		createPlan(t, f.router, f.token, fmt.Sprintf(
			`{"athlete_id":%q,"title":"Copy","start_date":"2026-10-23","source_plan_id":%q}`, athleteID, source), http.StatusCreated)
		return len(rec.take())
	}

	assert.Equal(t, copyOf(f.planID), copyOf(bigPlan))
}

// The insert helpers below bypass the builder endpoints for the same reason as insertPlan.

func insertMovement(t *testing.T, db *gorm.DB, name, category string) uuid.UUID {
	t.Helper()

	id := uuid.New()
	require.NoError(t, db.Create(&movement.Movement{ID: id, Name: name, Category: &category}).Error)

	return id
}

func insertDay(db *gorm.DB, planID uuid.UUID, label string, orderIndex int) (uuid.UUID, error) {
	id := uuid.New()
	err := db.Omit(clause.Associations).Create(&plan.Day{
		ID: id, PlanID: planID, Label: label, OrderIndex: orderIndex,
	}).Error

	return id, err
}

func insertBlock(t *testing.T, db *gorm.DB, dayID uuid.UUID, orderIndex, sets int, notes *string) uuid.UUID {
	t.Helper()

	rest := 60 + 30*orderIndex
	id := uuid.New()
	require.NoError(t, db.Omit(clause.Associations).Create(&plan.Block{
		ID: id, DayID: dayID, OrderIndex: orderIndex, Sets: sets, RestSeconds: &rest, Notes: notes,
	}).Error)

	return id
}

func insertBlockMovement(t *testing.T, db *gorm.DB, blockID, movementID uuid.UUID, reps, orderInBlock int) {
	t.Helper()

	require.NoError(t, db.Omit(clause.Associations).Create(&plan.BlockMovement{
		ID: uuid.New(), BlockID: blockID, MovementID: movementID, Reps: &reps, OrderInBlock: orderInBlock,
	}).Error)
}

// sqlRecorder captures every statement GORM runs, so tests can count queries instead of assuming Preload batches them.
type sqlRecorder struct {
	logger.Interface

	mu    sync.Mutex
	stmts []string
}

func recordSQL(db *gorm.DB) *sqlRecorder {
	r := &sqlRecorder{Interface: logger.Discard}
	db.Logger = r
	return r
}

func (r *sqlRecorder) Trace(_ context.Context, _ time.Time, fc func() (string, int64), _ error) {
	sql, _ := fc()

	r.mu.Lock()
	defer r.mu.Unlock()
	r.stmts = append(r.stmts, sql)
}

func (r *sqlRecorder) take() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := r.stmts
	r.stmts = nil
	return out
}

type planDetail struct {
	ID        string `json:"id"`
	AthleteID string `json:"athlete_id"`
	Title     string `json:"title"`
	StartDate string `json:"start_date"`
	Days      []struct {
		ID         string `json:"id"`
		Label      string `json:"label"`
		OrderIndex int    `json:"order_index"`
		Blocks     []struct {
			ID          string  `json:"id"`
			OrderIndex  int     `json:"order_index"`
			Sets        int     `json:"sets"`
			RestSeconds *int    `json:"rest_seconds"`
			Notes       *string `json:"notes"`
			Movements   []struct {
				ID           string  `json:"id"`
				MovementID   string  `json:"movement_id"`
				Name         string  `json:"name"`
				Category     *string `json:"category"`
				Reps         *int    `json:"reps"`
				Load         *string `json:"load"`
				OrderInBlock int     `json:"order_in_block"`
			} `json:"movements"`
		} `json:"blocks"`
	} `json:"days"`
}

func planPath(planID string) string {
	return "/api/v1/plans/" + planID
}

func getPlan(t *testing.T, router *gin.Engine, token, planID string) (planDetail, json.RawMessage) {
	t.Helper()

	env := decodeEnvelope(t, do(t, router, http.MethodGet, planPath(planID), token, ""), http.StatusOK)
	assert.True(t, env.Success)
	assert.Nil(t, env.Error)

	var out planDetail
	require.NoError(t, json.Unmarshal(env.Data, &out))

	return out, env.Data
}

func objectKeys(t *testing.T, raw json.RawMessage) []string {
	t.Helper()

	var obj map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &obj))

	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}

func TestPlanDetail_NestedTreeInDisplayOrder(t *testing.T) {
	router, db := newTestRouter(t)
	token := createCoach(t, router, db, "09121110301")
	athleteID := createAthlete(t, router, token, "09121230301")
	planID := insertPlan(t, db, athleteID, "Cut phase 1", daysFromToday(-10))

	squat := insertMovement(t, db, "Back squat", "strength")
	pullUp := insertMovement(t, db, "Pull-up", "strength")
	pushUp := insertMovement(t, db, "Push-up", "strength")
	deadlift := insertMovement(t, db, "Deadlift", "strength")
	rower := insertMovement(t, db, "Rower", "cardio")

	// Every level is inserted out of display order, so insertion order can't pass for ordering.
	dayB, err := insertDay(db, planID, "B", 1)
	require.NoError(t, err)
	dayA, err := insertDay(db, planID, "A", 0)
	require.NoError(t, err)

	superset := "superset"
	aSuperset := insertBlock(t, db, dayA, 1, 3, &superset)
	aSingle := insertBlock(t, db, dayA, 0, 4, nil)
	insertBlockMovement(t, db, aSuperset, pushUp, 15, 1)
	insertBlockMovement(t, db, aSuperset, pullUp, 10, 0)
	insertBlockMovement(t, db, aSingle, squat, 8, 0)

	bSecond := insertBlock(t, db, dayB, 1, 1, nil)
	bFirst := insertBlock(t, db, dayB, 0, 5, nil)
	insertBlockMovement(t, db, bSecond, rower, 20, 0)
	insertBlockMovement(t, db, bFirst, deadlift, 5, 0)

	rec := recordSQL(db)
	got, raw := getPlan(t, router, token, planID.String())
	stmts := rec.take()

	t.Run("plan fields", func(t *testing.T) {
		assert.Equal(t, planID.String(), got.ID)
		assert.Equal(t, athleteID, got.AthleteID)
		assert.Equal(t, "Cut phase 1", got.Title)
		assert.Equal(t, daysFromToday(-10).Format(time.DateOnly), got.StartDate)
	})

	t.Run("days, blocks and movements in order", func(t *testing.T) {
		require.Len(t, got.Days, 2)
		assert.Equal(t, "A", got.Days[0].Label)
		assert.Equal(t, 0, got.Days[0].OrderIndex)
		assert.Equal(t, "B", got.Days[1].Label)

		a := got.Days[0]
		require.Len(t, a.Blocks, 2)
		assert.Equal(t, []int{0, 1}, []int{a.Blocks[0].OrderIndex, a.Blocks[1].OrderIndex})
		assert.Equal(t, 4, a.Blocks[0].Sets)
		assert.Nil(t, a.Blocks[0].Notes)
		require.NotNil(t, a.Blocks[1].RestSeconds)
		assert.Equal(t, 90, *a.Blocks[1].RestSeconds)
		assert.Equal(t, &superset, a.Blocks[1].Notes)

		require.Len(t, a.Blocks[0].Movements, 1)
		assert.Equal(t, "Back squat", a.Blocks[0].Movements[0].Name)

		ss := a.Blocks[1].Movements
		require.Len(t, ss, 2, "a superset is one block holding both movements")
		assert.Equal(t, []string{"Pull-up", "Push-up"}, []string{ss[0].Name, ss[1].Name})
		assert.Equal(t, []int{0, 1}, []int{ss[0].OrderInBlock, ss[1].OrderInBlock})
		assert.Equal(t, pullUp.String(), ss[0].MovementID)
		require.NotNil(t, ss[0].Category)
		assert.Equal(t, "strength", *ss[0].Category)
		require.NotNil(t, ss[0].Reps)
		assert.Equal(t, 10, *ss[0].Reps)

		b := got.Days[1]
		require.Len(t, b.Blocks, 2)
		assert.Equal(t, "Deadlift", b.Blocks[0].Movements[0].Name)
		assert.Equal(t, "Rower", b.Blocks[1].Movements[0].Name)
		require.NotNil(t, b.Blocks[1].Movements[0].Category)
		assert.Equal(t, "cardio", *b.Blocks[1].Movements[0].Category)
	})

	t.Run("single and superset blocks share one shape", func(t *testing.T) {
		var tree struct {
			Days []struct {
				Blocks []json.RawMessage `json:"blocks"`
			} `json:"days"`
		}
		require.NoError(t, json.Unmarshal(raw, &tree))

		single, ss := tree.Days[0].Blocks[0], tree.Days[0].Blocks[1]
		assert.Equal(t, objectKeys(t, single), objectKeys(t, ss))

		var singleBlock, ssBlock struct {
			Movements []json.RawMessage `json:"movements"`
		}
		require.NoError(t, json.Unmarshal(single, &singleBlock))
		require.NoError(t, json.Unmarshal(ss, &ssBlock))
		assert.Len(t, singleBlock.Movements, 1)
		assert.Len(t, ssBlock.Movements, 2)
		assert.Equal(t, objectKeys(t, singleBlock.Movements[0]), objectKeys(t, ssBlock.Movements[0]))
	})

	t.Run("one query per level, not per row", func(t *testing.T) {
		for _, s := range stmts {
			t.Log(s)
		}

		// plan, athlete ownership check, then day, block, block_movement, movement: 2 days, 4 blocks and
		// 5 movements would take 13+ queries if any level were loaded per row.
		tables := make([]string, 0, len(stmts))
		for _, s := range stmts {
			require.True(t, strings.HasPrefix(s, "SELECT"), s)
			tables = append(tables, strings.Fields(s[strings.Index(s, " FROM ")+len(" FROM "):])[0])
		}
		// ElementsMatch: GORM logs a statement when it finishes, and nested preloads finish before their parents.
		assert.ElementsMatch(t, []string{`"plan"`, `"athlete"`, `"day"`, `"block"`, `"block_movement"`, `"movement"`}, tables)
	})

	t.Run("a soft-deleted movement still names itself", func(t *testing.T) {
		require.NoError(t, db.Delete(&movement.Movement{}, "id = ?", squat).Error)

		after, _ := getPlan(t, router, token, planID.String())
		assert.Equal(t, "Back squat", after.Days[0].Blocks[0].Movements[0].Name)
	})
}

func TestPlanDetail_EmptyLevelsAreEmptyArrays(t *testing.T) {
	router, db := newTestRouter(t)
	token := createCoach(t, router, db, "09121110302")
	athleteID := createAthlete(t, router, token, "09121230302")

	empty := insertPlan(t, db, athleteID, "Empty", daysFromToday(-1))
	_, raw := getPlan(t, router, token, empty.String())
	assert.JSONEq(t, `[]`, string(objectField(t, raw, "days")))

	withDay := insertPlan(t, db, athleteID, "One day", daysFromToday(-1))
	_, err := insertDay(db, withDay, "A", 0)
	require.NoError(t, err)

	got, _ := getPlan(t, router, token, withDay.String())
	require.Len(t, got.Days, 1)
	assert.NotNil(t, got.Days[0].Blocks)
	assert.Empty(t, got.Days[0].Blocks)
}

func objectField(t *testing.T, raw json.RawMessage, name string) json.RawMessage {
	t.Helper()

	var obj map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &obj))

	return obj[name]
}

func TestPlanDetail_AnotherCoachsPlanIsNotFound(t *testing.T) {
	router, db := newTestRouter(t)

	owner := createCoach(t, router, db, "09121110303")
	intruder := createCoach(t, router, db, "09121110304")
	athleteID := createAthlete(t, router, owner, "09121230303")
	planID := insertPlan(t, db, athleteID, "Base", daysFromToday(-30))

	t.Run("another coach gets 404, not 403", func(t *testing.T) {
		env := decodeEnvelope(t, do(t, router, http.MethodGet, planPath(planID.String()), intruder, ""), http.StatusNotFound)
		assert.False(t, env.Success)
		require.NotNil(t, env.Error)
		assert.Equal(t, httpx.CodeNotFound, env.Error.Code)
		assert.Equal(t, "null", string(env.Data), "the owner's plan doesn't leak")
	})

	t.Run("an id that doesn't exist reads the same", func(t *testing.T) {
		decodeEnvelope(t, do(t, router, http.MethodGet, planPath(uuid.NewString()), owner, ""), http.StatusNotFound)
	})

	t.Run("a malformed id reads the same", func(t *testing.T) {
		decodeEnvelope(t, do(t, router, http.MethodGet, planPath("not-a-uuid"), owner, ""), http.StatusNotFound)
	})

	t.Run("a soft-deleted athlete's plan reads the same", func(t *testing.T) {
		require.NoError(t, db.Delete(&athlete.Athlete{}, "id = ?", athleteID).Error)
		decodeEnvelope(t, do(t, router, http.MethodGet, planPath(planID.String()), owner, ""), http.StatusNotFound)
	})

	t.Run("no token is 401", func(t *testing.T) {
		require.Equal(t, http.StatusUnauthorized, do(t, router, http.MethodGet, planPath(planID.String()), "", "").Code)
	})
}

func TestDays_AtMostSevenPerPlan(t *testing.T) {
	router, db := newTestRouter(t)
	token := createCoach(t, router, db, "09121110305")
	athleteID := createAthlete(t, router, token, "09121230305")
	planID := insertPlan(t, db, athleteID, "Full week", daysFromToday(-1))

	for i := range plan.MaxDaysPerPlan {
		_, err := insertDay(db, planID, fmt.Sprintf("day%d", i+1), i)
		require.NoError(t, err)
	}

	_, err := insertDay(db, planID, "extra", plan.MaxDaysPerPlan)
	assert.ErrorIs(t, err, gorm.ErrCheckConstraintViolated, "an 8th day has no order_index slot left")

	_, err = insertDay(db, planID, "dup", 3)
	assert.ErrorIs(t, err, gorm.ErrDuplicatedKey, "two days can't share a slot")

	_, err = insertDay(db, planID, "negative", -1)
	assert.ErrorIs(t, err, gorm.ErrCheckConstraintViolated)

	got, _ := getPlan(t, router, token, planID.String())
	assert.Len(t, got.Days, plan.MaxDaysPerPlan)

	other := insertPlan(t, db, athleteID, "Another week", daysFromToday(1))
	_, err = insertDay(db, other, "A", 0)
	assert.NoError(t, err, "the cap is per plan")
}

func coachIDByPhone(t *testing.T, db *gorm.DB, phone string) uuid.UUID {
	t.Helper()

	var c coach.Coach
	require.NoError(t, db.First(&c, "phone = ?", phone).Error)

	return c.ID
}

func insertCoachMovement(t *testing.T, db *gorm.DB, coachID uuid.UUID, name string) uuid.UUID {
	t.Helper()

	id := uuid.New()
	require.NoError(t, db.Create(&movement.Movement{ID: id, CoachID: &coachID, Name: name}).Error)

	return id
}

func dataID(t *testing.T, env envelope) string {
	t.Helper()

	var created struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(env.Data, &created))
	require.NotEmpty(t, created.ID)

	return created.ID
}

func daysPath(planID string) string       { return planPath(planID) + "/days" }
func blocksPath(dayID string) string      { return "/api/v1/days/" + dayID + "/blocks" }
func movementsPath(blockID string) string { return "/api/v1/blocks/" + blockID + "/movements" }

type builtTree struct {
	planID, dayID, blockID string
}

// buildToBlock goes through the builder endpoints, leaving an empty block to add movements to.
func buildToBlock(t *testing.T, router *gin.Engine, token, athleteID string) builtTree {
	t.Helper()

	var tree builtTree

	body := fmt.Sprintf(`{"athlete_id":%q,"start_date":"2026-09-15","title":"Cut phase 1"}`, athleteID)
	tree.planID = dataID(t, decodeEnvelope(t, do(t, router, http.MethodPost, "/api/v1/plans", token, body), http.StatusCreated))

	env := decodeEnvelope(t, do(t, router, http.MethodPost, daysPath(tree.planID), token, `{"label":"A","order_index":0}`), http.StatusCreated)
	tree.dayID = dataID(t, env)

	env = decodeEnvelope(t, do(t, router, http.MethodPost, blocksPath(tree.dayID), token, `{"order_index":0,"sets":3}`), http.StatusCreated)
	tree.blockID = dataID(t, env)

	return tree
}

func countBlockMovements(t *testing.T, db *gorm.DB, blockID string) int64 {
	t.Helper()

	var n int64
	require.NoError(t, db.Model(&plan.BlockMovement{}).Where("block_id = ?", blockID).Count(&n).Error)

	return n
}

func requireValidationFields(t *testing.T, rec *httptest.ResponseRecorder) map[string]string {
	t.Helper()

	env := decodeEnvelope(t, rec, http.StatusBadRequest)
	assert.False(t, env.Success)
	assert.Equal(t, "null", string(env.Data))
	require.NotNil(t, env.Error)
	assert.Equal(t, httpx.CodeValidationFailed, env.Error.Code)

	return env.Error.Fields
}

func TestBuilder_FullPlanBuildUp(t *testing.T) {
	router, db := newTestRouter(t)
	const coachPhone = "09121110401"
	token := createCoach(t, router, db, coachPhone)
	athleteID := createAthlete(t, router, token, "09121230401")

	// One universal and one of the coach's own: a superset can draw from both halves of the library.
	pullUp := insertMovement(t, db, "Pull-up", "strength")
	ringDip := insertCoachMovement(t, db, coachIDByPhone(t, db, coachPhone), "Ring dip")

	createBody := fmt.Sprintf(`{"athlete_id":%q,"start_date":"2026-09-15","title":" Cut phase 1 ","note":"Deload every 4th week"}`, athleteID)
	env := decodeEnvelope(t, do(t, router, http.MethodPost, "/api/v1/plans", token, createBody), http.StatusCreated)
	assert.True(t, env.Success)
	assert.Nil(t, env.Error)

	var createdPlan struct {
		ID        string          `json:"id"`
		AthleteID string          `json:"athlete_id"`
		Title     string          `json:"title"`
		Note      *string         `json:"note"`
		StartDate string          `json:"start_date"`
		Days      json.RawMessage `json:"days"`
	}
	require.NoError(t, json.Unmarshal(env.Data, &createdPlan))
	assert.Equal(t, athleteID, createdPlan.AthleteID)
	assert.Equal(t, "Cut phase 1", createdPlan.Title, "title is trimmed")
	require.NotNil(t, createdPlan.Note)
	assert.Equal(t, "Deload every 4th week", *createdPlan.Note)
	assert.Equal(t, "2026-09-15", createdPlan.StartDate)
	assert.JSONEq(t, `[]`, string(createdPlan.Days))

	env = decodeEnvelope(t, do(t, router, http.MethodPost, daysPath(createdPlan.ID), token, `{"label":"A","order_index":0}`), http.StatusCreated)
	var createdDay struct {
		ID         string          `json:"id"`
		Label      string          `json:"label"`
		OrderIndex int             `json:"order_index"`
		Blocks     json.RawMessage `json:"blocks"`
	}
	require.NoError(t, json.Unmarshal(env.Data, &createdDay))
	assert.NotEmpty(t, createdDay.ID)
	assert.Equal(t, "A", createdDay.Label)
	assert.Equal(t, 0, createdDay.OrderIndex)
	assert.JSONEq(t, `[]`, string(createdDay.Blocks))

	blockBody := `{"order_index":0,"sets":3,"rest_seconds":90,"notes":"superset"}`
	env = decodeEnvelope(t, do(t, router, http.MethodPost, blocksPath(createdDay.ID), token, blockBody), http.StatusCreated)
	var createdBlock struct {
		ID          string          `json:"id"`
		OrderIndex  int             `json:"order_index"`
		Sets        int             `json:"sets"`
		RestSeconds *int            `json:"rest_seconds"`
		Notes       *string         `json:"notes"`
		Movements   json.RawMessage `json:"movements"`
	}
	require.NoError(t, json.Unmarshal(env.Data, &createdBlock))
	assert.NotEmpty(t, createdBlock.ID)
	assert.Equal(t, 3, createdBlock.Sets)
	require.NotNil(t, createdBlock.RestSeconds)
	assert.Equal(t, 90, *createdBlock.RestSeconds)
	require.NotNil(t, createdBlock.Notes)
	assert.Equal(t, "superset", *createdBlock.Notes)
	assert.JSONEq(t, `[]`, string(createdBlock.Movements))

	supersetBody := fmt.Sprintf(`[
		{"movement_id":%q,"reps":10,"load":"bodyweight","order_in_block":0},
		{"movement_id":%q,"reps":12,"order_in_block":1}
	]`, pullUp, ringDip)
	env = decodeEnvelope(t, do(t, router, http.MethodPost, movementsPath(createdBlock.ID), token, supersetBody), http.StatusCreated)
	var createdMovements []struct {
		ID           string  `json:"id"`
		MovementID   string  `json:"movement_id"`
		Name         string  `json:"name"`
		Category     *string `json:"category"`
		Reps         *int    `json:"reps"`
		Load         *string `json:"load"`
		OrderInBlock int     `json:"order_in_block"`
	}
	require.NoError(t, json.Unmarshal(env.Data, &createdMovements))
	require.Len(t, createdMovements, 2)
	assert.Equal(t, []string{"Pull-up", "Ring dip"}, []string{createdMovements[0].Name, createdMovements[1].Name},
		"the response names each movement from the library")
	assert.Equal(t, pullUp.String(), createdMovements[0].MovementID)
	require.NotNil(t, createdMovements[0].Category)
	assert.Equal(t, "strength", *createdMovements[0].Category)
	require.NotNil(t, createdMovements[0].Load)
	assert.Equal(t, "bodyweight", *createdMovements[0].Load)
	assert.Nil(t, createdMovements[1].Load)
	require.NotNil(t, createdMovements[1].Reps)
	assert.Equal(t, 12, *createdMovements[1].Reps)

	got, _ := getPlan(t, router, token, createdPlan.ID)
	assert.Equal(t, "Cut phase 1", got.Title)
	require.Len(t, got.Days, 1)
	assert.Equal(t, createdDay.ID, got.Days[0].ID)
	require.Len(t, got.Days[0].Blocks, 1)
	block := got.Days[0].Blocks[0]
	assert.Equal(t, createdBlock.ID, block.ID)
	assert.Equal(t, 3, block.Sets)
	require.Len(t, block.Movements, 2, "both movements landed in one block: a superset")
	assert.Equal(t, []string{createdMovements[0].ID, createdMovements[1].ID}, []string{block.Movements[0].ID, block.Movements[1].ID})
	assert.Equal(t, []int{0, 1}, []int{block.Movements[0].OrderInBlock, block.Movements[1].OrderInBlock})
	require.NotNil(t, block.Movements[0].Load)
	assert.Equal(t, "bodyweight", *block.Movements[0].Load)

	listed := listPlans(t, router, token, athleteID)
	require.Len(t, listed, 1)
	assert.Equal(t, createdPlan.ID, listed[0].ID)
}

func TestBuilder_MovementOutsideLibraryRejectsWholeBatch(t *testing.T) {
	router, db := newTestRouter(t)
	const coachPhone, otherPhone = "09121110402", "09121110403"
	token := createCoach(t, router, db, coachPhone)
	createCoach(t, router, db, otherPhone)
	athleteID := createAthlete(t, router, token, "09121230402")
	tree := buildToBlock(t, router, token, athleteID)

	squat := insertMovement(t, db, "Back squat", "strength")
	ownLunge := insertCoachMovement(t, db, coachIDByPhone(t, db, coachPhone), "Walking lunge")
	othersLunge := insertCoachMovement(t, db, coachIDByPhone(t, db, otherPhone), "Walking lunge")
	retired := insertCoachMovement(t, db, coachIDByPhone(t, db, coachPhone), "Retired move")
	require.NoError(t, db.Delete(&movement.Movement{}, "id = ?", retired).Error)

	batch := func(second string) string {
		return fmt.Sprintf(`[{"movement_id":%q,"reps":8,"order_in_block":0},{"movement_id":%q,"reps":10,"order_in_block":1}]`, squat, second)
	}

	for name, second := range map[string]string{
		"another coach's movement":      othersLunge.String(),
		"a movement that doesn't exist": uuid.NewString(),
		"a soft-deleted movement":       retired.String(),
	} {
		t.Run(name, func(t *testing.T) {
			fields := requireValidationFields(t, do(t, router, http.MethodPost, movementsPath(tree.blockID), token, batch(second)))

			assert.Equal(t, map[string]string{"[1].movement_id": "not found"}, fields,
				"the bad entry is pinpointed, and reads the same whether it's foreign, missing or deleted")
			assert.Zero(t, countBlockMovements(t, db, tree.blockID), "the valid first movement wasn't inserted either")
		})
	}

	t.Run("every bad entry is reported at once", func(t *testing.T) {
		body := fmt.Sprintf(`[{"movement_id":%q,"order_in_block":0},{"movement_id":"not-a-uuid","order_in_block":1},{"movement_id":%q,"reps":0}]`,
			othersLunge, squat)

		fields := requireValidationFields(t, do(t, router, http.MethodPost, movementsPath(tree.blockID), token, body))

		assert.Equal(t, map[string]string{
			"[0].movement_id":    "not found",
			"[1].movement_id":    "must be a UUID",
			"[2].reps":           "must be between 1 and 1000",
			"[2].order_in_block": httpx.MsgFieldRequired,
		}, fields)
		assert.Zero(t, countBlockMovements(t, db, tree.blockID))
	})

	t.Run("an empty batch is rejected", func(t *testing.T) {
		env := decodeEnvelope(t, do(t, router, http.MethodPost, movementsPath(tree.blockID), token, `[]`), http.StatusBadRequest)
		require.NotNil(t, env.Error)
		assert.Equal(t, httpx.CodeValidationFailed, env.Error.Code)
	})

	t.Run("the same batch with the coach's own movement goes through", func(t *testing.T) {
		decodeEnvelope(t, do(t, router, http.MethodPost, movementsPath(tree.blockID), token, batch(ownLunge.String())), http.StatusCreated)
		assert.Equal(t, int64(2), countBlockMovements(t, db, tree.blockID))
	})
}

// The service rejects bad movements before inserting, so this drives the repository directly to fail
// mid-batch. One row per INSERT and no GORM default transaction leave only the explicit one to roll back.
func TestCreateBlockMovements_MidBatchFailureInsertsNothing(t *testing.T) {
	router, db := newTestRouter(t)
	token := createCoach(t, router, db, "09121110404")
	athleteID := createAthlete(t, router, token, "09121230404")
	tree := buildToBlock(t, router, token, athleteID)
	squat := insertMovement(t, db, "Back squat", "strength")

	repo := plan.NewRepository(db.Session(&gorm.Session{CreateBatchSize: 1, SkipDefaultTransaction: true}))
	blockID := uuid.MustParse(tree.blockID)

	err := repo.CreateBlockMovements(context.Background(), []plan.BlockMovement{
		{ID: uuid.New(), BlockID: blockID, MovementID: squat, OrderInBlock: 0},
		{ID: uuid.New(), BlockID: blockID, MovementID: uuid.New(), OrderInBlock: 1},
	})

	require.ErrorIs(t, err, gorm.ErrForeignKeyViolated, "the second row names no movement")
	assert.Zero(t, countBlockMovements(t, db, tree.blockID), "the first row was rolled back with it")
}

func TestBuilder_DayLabelAndSlotRules(t *testing.T) {
	router, db := newTestRouter(t)
	token := createCoach(t, router, db, "09121110405")
	athleteID := createAthlete(t, router, token, "09121230405")

	body := fmt.Sprintf(`{"athlete_id":%q,"start_date":"2026-09-15","title":"Week"}`, athleteID)
	planID := dataID(t, decodeEnvelope(t, do(t, router, http.MethodPost, "/api/v1/plans", token, body), http.StatusCreated))

	for name, tc := range map[string]struct {
		body   string
		fields map[string]string
	}{
		"unknown label":           {`{"label":"Z","order_index":0}`, map[string]string{"label": "must be one of A-G or day1-day7"}},
		"label past the 7th slot": {`{"label":"day8","order_index":0}`, map[string]string{"label": "must be one of A-G or day1-day7"}},
		"lowercase letter":        {`{"label":"a","order_index":0}`, map[string]string{"label": "must be one of A-G or day1-day7"}},
		"slot past the 7th":       {`{"label":"A","order_index":7}`, map[string]string{"order_index": "must be between 0 and 6"}},
		"negative slot":           {`{"label":"A","order_index":-1}`, map[string]string{"order_index": "must be between 0 and 6"}},
		"slot beyond INT":         {`{"label":"A","order_index":9999999999}`, map[string]string{"order_index": "must be between 0 and 6"}},
		"nothing given": {`{}`, map[string]string{
			"label":       httpx.MsgFieldRequired,
			"order_index": httpx.MsgFieldRequired,
		}},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.fields, requireValidationFields(t, do(t, router, http.MethodPost, daysPath(planID), token, tc.body)))
		})
	}

	decodeEnvelope(t, do(t, router, http.MethodPost, daysPath(planID), token, `{"label":"A","order_index":0}`), http.StatusCreated)

	t.Run("a taken slot is a 400 on order_index, not a 500", func(t *testing.T) {
		fields := requireValidationFields(t, do(t, router, http.MethodPost, daysPath(planID), token, `{"label":"B","order_index":0}`))
		assert.Equal(t, map[string]string{"order_index": "is already used by another day in this plan"}, fields)
	})

	t.Run("the schema's range check maps the same way if it's ever reached", func(t *testing.T) {
		err := plan.NewRepository(db).CreateDay(context.Background(), &plan.Day{
			ID: uuid.New(), PlanID: uuid.MustParse(planID), Label: "H", OrderIndex: plan.MaxDaysPerPlan,
		})
		assert.ErrorIs(t, err, plan.ErrDaySlotOutOfRange)
	})

	got, _ := getPlan(t, router, token, planID)
	require.Len(t, got.Days, 1, "no rejected day was stored")
	assert.Equal(t, "A", got.Days[0].Label)
}

func TestBuilder_AnotherCoachsTreeIsNotFound(t *testing.T) {
	router, db := newTestRouter(t)
	owner := createCoach(t, router, db, "09121110406")
	intruder := createCoach(t, router, db, "09121110407")
	athleteID := createAthlete(t, router, owner, "09121230406")
	tree := buildToBlock(t, router, owner, athleteID)
	squat := insertMovement(t, db, "Back squat", "strength")

	// Every body is valid, and squat is universal, so only ownership can turn these away.
	createPlan := fmt.Sprintf(`{"athlete_id":%q,"start_date":"2026-09-15","title":"Stolen"}`, athleteID)
	addDay := `{"label":"B","order_index":1}`
	addBlock := `{"order_index":1,"sets":3}`
	addMovements := fmt.Sprintf(`[{"movement_id":%q,"reps":8,"order_in_block":0}]`, squat)

	requireNotFound := func(t *testing.T, target, token, body string) {
		t.Helper()

		env := decodeEnvelope(t, do(t, router, http.MethodPost, target, token, body), http.StatusNotFound)
		assert.False(t, env.Success)
		require.NotNil(t, env.Error)
		assert.Equal(t, httpx.CodeNotFound, env.Error.Code)
		assert.Equal(t, "null", string(env.Data))
	}

	t.Run("another coach gets 404 at every level, not 403", func(t *testing.T) {
		requireNotFound(t, "/api/v1/plans", intruder, createPlan)
		requireNotFound(t, daysPath(tree.planID), intruder, addDay)
		requireNotFound(t, blocksPath(tree.dayID), intruder, addBlock)
		requireNotFound(t, movementsPath(tree.blockID), intruder, addMovements)
	})

	t.Run("nothing was added to the owner's tree", func(t *testing.T) {
		assert.Len(t, listPlans(t, router, owner, athleteID), 1)

		got, _ := getPlan(t, router, owner, tree.planID)
		require.Len(t, got.Days, 1)
		require.Len(t, got.Days[0].Blocks, 1)
		assert.Empty(t, got.Days[0].Blocks[0].Movements)
	})

	t.Run("ids that don't exist read the same", func(t *testing.T) {
		missingAthlete := fmt.Sprintf(`{"athlete_id":%q,"start_date":"2026-09-15","title":"Ghost"}`, uuid.NewString())
		requireNotFound(t, "/api/v1/plans", owner, missingAthlete)
		requireNotFound(t, daysPath(uuid.NewString()), owner, addDay)
		requireNotFound(t, blocksPath(uuid.NewString()), owner, addBlock)
		requireNotFound(t, movementsPath(uuid.NewString()), owner, addMovements)
	})

	t.Run("malformed path ids read the same", func(t *testing.T) {
		requireNotFound(t, daysPath("not-a-uuid"), owner, addDay)
		requireNotFound(t, blocksPath("not-a-uuid"), owner, addBlock)
		requireNotFound(t, movementsPath("not-a-uuid"), owner, addMovements)
	})

	t.Run("no token is 401", func(t *testing.T) {
		require.Equal(t, http.StatusUnauthorized, do(t, router, http.MethodPost, movementsPath(tree.blockID), "", addMovements).Code)
	})

	t.Run("a soft-deleted athlete's tree reads the same", func(t *testing.T) {
		require.NoError(t, db.Delete(&athlete.Athlete{}, "id = ?", athleteID).Error)

		requireNotFound(t, "/api/v1/plans", owner, createPlan)
		requireNotFound(t, daysPath(tree.planID), owner, addDay)
		requireNotFound(t, blocksPath(tree.dayID), owner, addBlock)
		requireNotFound(t, movementsPath(tree.blockID), owner, addMovements)
	})
}

func TestBuilder_CreatePlanRejectsBadInputWithFieldErrors(t *testing.T) {
	router, db := newTestRouter(t)
	token := createCoach(t, router, db, "09121110408")

	fields := requireValidationFields(t, do(t, router, http.MethodPost, "/api/v1/plans", token,
		`{"athlete_id":"not-a-uuid","start_date":"15/09/2026","title":"   "}`))

	assert.Equal(t, map[string]string{
		"athlete_id": "must be a UUID",
		"start_date": httpx.MsgDateFormat,
		"title":      httpx.MsgFieldRequired,
	}, fields)
}
