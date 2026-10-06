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
	"github.com/AbolfazlTayyar/gogym/internal/httpx"
	"github.com/AbolfazlTayyar/gogym/internal/models"
	"github.com/AbolfazlTayyar/gogym/internal/plan"
	"github.com/AbolfazlTayyar/gogym/internal/testutil"
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

// The insert helpers below write straight to the tables until the plan-builder endpoints exist.

func insertMovement(t *testing.T, db *gorm.DB, name, category string) uuid.UUID {
	t.Helper()

	id := uuid.New()
	require.NoError(t, db.Create(&models.Movement{ID: id, Name: name, Category: &category}).Error)

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
		Label      string `json:"label"`
		OrderIndex int    `json:"order_index"`
		Blocks     []struct {
			OrderIndex  int     `json:"order_index"`
			Sets        int     `json:"sets"`
			RestSeconds *int    `json:"rest_seconds"`
			Notes       *string `json:"notes"`
			Movements   []struct {
				MovementID   string  `json:"movement_id"`
				Name         string  `json:"name"`
				Category     *string `json:"category"`
				Reps         *int    `json:"reps"`
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
	token := signupCoach(t, router, "09121110301")
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
		require.NoError(t, db.Delete(&models.Movement{}, "id = ?", squat).Error)

		after, _ := getPlan(t, router, token, planID.String())
		assert.Equal(t, "Back squat", after.Days[0].Blocks[0].Movements[0].Name)
	})
}

func TestPlanDetail_EmptyLevelsAreEmptyArrays(t *testing.T) {
	router, db := newTestRouter(t)
	token := signupCoach(t, router, "09121110302")
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

	owner := signupCoach(t, router, "09121110303")
	intruder := signupCoach(t, router, "09121110304")
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
	token := signupCoach(t, router, "09121110305")
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
