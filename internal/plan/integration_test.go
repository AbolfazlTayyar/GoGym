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
	"github.com/AbolfazlTayyar/gogym/internal/movement"
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
	planSvc := plan.NewService(plan.NewRepository(db), athleteSvc, movement.NewService(movement.NewRepository(db)))

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
		require.NoError(t, db.Delete(&movement.Movement{}, "id = ?", squat).Error)

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
	token := signupCoach(t, router, coachPhone)
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
	token := signupCoach(t, router, coachPhone)
	signupCoach(t, router, otherPhone)
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
	token := signupCoach(t, router, "09121110404")
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
	token := signupCoach(t, router, "09121110405")
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
	owner := signupCoach(t, router, "09121110406")
	intruder := signupCoach(t, router, "09121110407")
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
	router, _ := newTestRouter(t)
	token := signupCoach(t, router, "09121110408")

	fields := requireValidationFields(t, do(t, router, http.MethodPost, "/api/v1/plans", token,
		`{"athlete_id":"not-a-uuid","start_date":"15/09/2026","title":"   "}`))

	assert.Equal(t, map[string]string{
		"athlete_id": "must be a UUID",
		"start_date": httpx.MsgDateFormat,
		"title":      httpx.MsgFieldRequired,
	}, fields)
}
