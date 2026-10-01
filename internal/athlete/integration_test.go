package athlete_test

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
	"github.com/AbolfazlTayyar/gogym/internal/testutil"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const athletesPath = "/api/v1/athletes"

// newTestRouter wires the coach module too, since athlete requests need a real coach's token.
func newTestRouter(t *testing.T) *gin.Engine {
	t.Helper()

	router, _ := newTestRouterWithDB(t)

	return router
}

func newTestRouterWithDB(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()

	db := testutil.NewDB(t)

	coachRepo := coach.NewRepository(db)
	coachSvc := coach.NewService(coachRepo, "test-secret", time.Hour)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	v1 := router.Group("/api/v1")
	protected := v1.Group("")
	protected.Use(coach.AuthMiddleware(coachSvc))

	coach.RegisterRoutes(v1, protected, coach.NewHandler(coachSvc, coachRepo))
	athlete.RegisterRoutes(protected, athlete.NewHandler(athlete.NewService(athlete.NewRepository(db))))

	return router, db
}

type athleteResponse struct {
	ID              string   `json:"id"`
	FirstName       string   `json:"first_name"`
	LastName        string   `json:"last_name"`
	Phone           string   `json:"phone"`
	ExperienceLevel *string  `json:"experience_level"`
	Height          *float64 `json:"height"`
	AthleteType     string   `json:"athlete_type"`
}

type listMeta struct {
	Total  int64 `json:"total"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}

type envelope struct {
	Success bool             `json:"success"`
	Data    json.RawMessage  `json:"data"`
	Error   *httpx.ErrorBody `json:"error"`
	Meta    json.RawMessage  `json:"meta"`
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

	login := do(t, router, http.MethodPost, "/api/v1/auth/login", "", credentials)
	env := decodeEnvelope(t, login, http.StatusOK)

	var payload struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.Unmarshal(env.Data, &payload))
	require.NotEmpty(t, payload.Token)

	return payload.Token
}

func createAthlete(t *testing.T, router *gin.Engine, token, body string) athleteResponse {
	t.Helper()

	rec := do(t, router, http.MethodPost, athletesPath, token, body)
	env := decodeEnvelope(t, rec, http.StatusCreated)
	assert.True(t, env.Success)
	assert.Nil(t, env.Error)

	var created athleteResponse
	require.NoError(t, json.Unmarshal(env.Data, &created))

	return created
}

func listAthletes(t *testing.T, router *gin.Engine, token, query string) ([]athleteResponse, listMeta) {
	t.Helper()

	rec := do(t, router, http.MethodGet, athletesPath+query, token, "")
	env := decodeEnvelope(t, rec, http.StatusOK)
	assert.True(t, env.Success)
	assert.Nil(t, env.Error)

	var athletes []athleteResponse
	require.NoError(t, json.Unmarshal(env.Data, &athletes))

	var meta listMeta
	require.NoError(t, json.Unmarshal(env.Meta, &meta))

	return athletes, meta
}

func names(athletes []athleteResponse) []string {
	out := make([]string, 0, len(athletes))
	for _, a := range athletes {
		out = append(out, a.FirstName)
	}
	return out
}

func TestAthleteLifecycle(t *testing.T) {
	router := newTestRouter(t)

	coachA := signupCoach(t, router, "09121110001")
	coachB := signupCoach(t, router, "09121110002")

	created := createAthlete(t, router, coachA, `{
		"first_name": "Sara", "last_name": "Ahmadi", "phone": "09121230001",
		"experience_level": "beginner", "goal": "lose 5kg", "height": 168
	}`)
	require.NotEmpty(t, created.ID)
	assert.Equal(t, athlete.TypePrivate, created.AthleteType, "athlete_type defaults to private")
	require.NotNil(t, created.ExperienceLevel)
	assert.Equal(t, athlete.ExperienceBeginner, *created.ExperienceLevel)

	createAthlete(t, router, coachA, `{"first_name":"Reza","last_name":"Karimi","phone":"09121230002"}`)

	// Coach B's matching name makes a cross-tenant leak visible in the search below.
	createAthlete(t, router, coachB, `{"first_name":"Sara","last_name":"Beheshti","phone":"09121230003"}`)

	t.Run("created athletes appear in the coach's list", func(t *testing.T) {
		athletes, meta := listAthletes(t, router, coachA, "")

		assert.ElementsMatch(t, []string{"Sara", "Reza"}, names(athletes))
		assert.Equal(t, int64(2), meta.Total)
		assert.Equal(t, 20, meta.Limit, "the default page size is reported back")
		assert.Equal(t, 0, meta.Offset)
	})

	t.Run("search matches a partial name, case-insensitively", func(t *testing.T) {
		athletes, meta := listAthletes(t, router, coachA, "?q=sar")

		require.Len(t, athletes, 1)
		assert.Equal(t, "Sara", athletes[0].FirstName)
		assert.Equal(t, "Ahmadi", athletes[0].LastName, "coach B's Sara is not in coach A's results")
		assert.Equal(t, int64(1), meta.Total)
	})

	t.Run("search matches the last name too", func(t *testing.T) {
		athletes, _ := listAthletes(t, router, coachA, "?q=KARIM")

		require.Len(t, athletes, 1)
		assert.Equal(t, "Reza", athletes[0].FirstName)
	})

	t.Run("another coach's athletes never appear", func(t *testing.T) {
		athletes, meta := listAthletes(t, router, coachB, "")

		require.Len(t, athletes, 1)
		assert.Equal(t, "Beheshti", athletes[0].LastName)
		assert.Equal(t, int64(1), meta.Total)
	})

	t.Run("a wildcard in the search is matched literally", func(t *testing.T) {
		athletes, meta := listAthletes(t, router, coachA, "?q=%25")

		assert.Empty(t, athletes, "a bare %% must not match every athlete")
		assert.Equal(t, int64(0), meta.Total)
	})

	t.Run("the list is rejected without a token", func(t *testing.T) {
		rec := do(t, router, http.MethodGet, athletesPath, "", "")
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestList_AthleteTypeFilter(t *testing.T) {
	router := newTestRouter(t)
	token := signupCoach(t, router, "09121110003")

	createAthlete(t, router, token, `{"first_name":"Private","last_name":"One","phone":"09121230011"}`)
	createAthlete(t, router, token, `{"first_name":"Public","last_name":"Two","phone":"09121230012","athlete_type":"public"}`)

	t.Run("defaults to private only", func(t *testing.T) {
		athletes, meta := listAthletes(t, router, token, "")

		require.Len(t, athletes, 1)
		assert.Equal(t, athlete.TypePrivate, athletes[0].AthleteType)
		assert.Equal(t, int64(1), meta.Total, "the total counts the filtered set, not the roster")
	})

	t.Run("public athletes are reachable, not hidden", func(t *testing.T) {
		athletes, _ := listAthletes(t, router, token, "?athlete_type=public")

		require.Len(t, athletes, 1)
		assert.Equal(t, "Public", athletes[0].FirstName)
	})

	t.Run("all returns both", func(t *testing.T) {
		athletes, meta := listAthletes(t, router, token, "?athlete_type=all")

		assert.ElementsMatch(t, []string{"Private", "Public"}, names(athletes))
		assert.Equal(t, int64(2), meta.Total)
	})

	t.Run("the filter and the search compose", func(t *testing.T) {
		athletes, _ := listAthletes(t, router, token, "?athlete_type=all&q=pub")

		require.Len(t, athletes, 1)
		assert.Equal(t, "Public", athletes[0].FirstName)
	})
}

func TestList_Paginates(t *testing.T) {
	router := newTestRouter(t)
	token := signupCoach(t, router, "09121110004")

	const total = 5
	for i := range total {
		createAthlete(t, router, token, fmt.Sprintf(
			`{"first_name":"Athlete%d","last_name":"Test","phone":"091212300%02d"}`, i, i+20))
	}

	first, meta := listAthletes(t, router, token, "?limit=2")
	require.Len(t, first, 2)
	assert.Equal(t, int64(total), meta.Total, "the total is the whole match, not the page")
	assert.Equal(t, 2, meta.Limit)

	second, meta := listAthletes(t, router, token, "?limit=2&offset=2")
	require.Len(t, second, 2)
	assert.Equal(t, 2, meta.Offset)

	last, _ := listAthletes(t, router, token, "?limit=2&offset=4")
	require.Len(t, last, 1, "the final page is short")

	seen := append(append(names(first), names(second)...), names(last)...)
	assert.Len(t, seen, total)
	assert.ElementsMatch(t, []string{"Athlete0", "Athlete1", "Athlete2", "Athlete3", "Athlete4"}, seen,
		"the pages partition the roster with no overlap or gap")
}

func TestCreate_RejectsBadInputWithFieldErrors(t *testing.T) {
	router := newTestRouter(t)
	token := signupCoach(t, router, "09121110005")

	rec := do(t, router, http.MethodPost, athletesPath, token,
		`{"first_name":"Sara","last_name":"","phone":"+15550001111","athlete_type":"vip"}`)

	env := decodeEnvelope(t, rec, http.StatusBadRequest)
	assert.False(t, env.Success)
	assert.Equal(t, "null", string(env.Data))
	require.NotNil(t, env.Error)
	assert.Equal(t, httpx.CodeValidationFailed, env.Error.Code)
	assert.Contains(t, env.Error.Fields, "last_name")
	assert.Contains(t, env.Error.Fields, "phone")
	assert.Contains(t, env.Error.Fields, "athlete_type")

	athletes, _ := listAthletes(t, router, token, "")
	assert.Empty(t, athletes, "a rejected create writes nothing")
}

func TestCreate_IgnoresCoachIDInBody(t *testing.T) {
	router := newTestRouter(t)

	victim := signupCoach(t, router, "09121110006")
	attacker := signupCoach(t, router, "09121110007")

	createAthlete(t, router, victim, `{"first_name":"Sara","last_name":"Ahmadi","phone":"09121230031"}`)

	// Plant the victim's real coach id, so the test fails if the body is trusted.
	meEnv := decodeEnvelope(t, do(t, router, http.MethodGet, "/api/v1/coaches/me", victim, ""), http.StatusOK)
	var me struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(meEnv.Data, &me))

	createAthlete(t, router, attacker, fmt.Sprintf(
		`{"first_name":"Planted","last_name":"Row","phone":"09121230032","coach_id":%q}`, me.ID))

	athletes, _ := listAthletes(t, router, victim, "")
	assert.Equal(t, []string{"Sara"}, names(athletes), "the planted row landed under the attacker, not the victim")
}

// GORM's soft-delete scope comes from the model, so a Count without one would include deleted rows.
func TestList_ExcludesSoftDeletedAthletes(t *testing.T) {
	router, db := newTestRouterWithDB(t)
	token := signupCoach(t, router, "09121110008")

	kept := createAthlete(t, router, token, `{"first_name":"Kept","last_name":"Athlete","phone":"09121230041"}`)
	removed := createAthlete(t, router, token, `{"first_name":"Removed","last_name":"Athlete","phone":"09121230042"}`)

	require.NoError(t, db.Delete(&athlete.Athlete{}, "id = ?", removed.ID).Error)

	athletes, meta := listAthletes(t, router, token, "")

	require.Len(t, athletes, 1)
	assert.Equal(t, kept.ID, athletes[0].ID)
	assert.Equal(t, int64(1), meta.Total, "the total excludes the soft-deleted row too")

	var stillStored int64
	require.NoError(t, db.Unscoped().Model(&athlete.Athlete{}).
		Where("id = ? AND deleted_at IS NOT NULL", removed.ID).Count(&stillStored).Error)
	assert.Equal(t, int64(1), stillStored, "soft delete preserves the row")
}
