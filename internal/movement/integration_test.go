package movement_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"slices"
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
)

const movementsPath = "/api/v1/movements"

// seededBackSquat is one of the universal movements the seed migration inserts.
const (
	seededBackSquat     = "e3935578-9941-45f9-b273-4ce0086cb0b8"
	seededBackSquatName = "Back squat"
)

const seededCount = 20

const legs = "legs"

// newTestRouter wires plan too, so a test can check what a deleted movement does to a plan that uses it.
func newTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()

	db := testutil.NewDB(t)

	coachRepo := coach.NewRepository(db)
	coachSvc := coach.NewService(coachRepo, "test-secret", time.Hour)
	athleteSvc := athlete.NewService(athlete.NewRepository(db))

	gin.SetMode(gin.TestMode)
	router := gin.New()

	v1 := router.Group("/api/v1")
	protected := v1.Group("")
	protected.Use(coach.AuthMiddleware(coachSvc))

	coach.RegisterRoutes(v1, protected, coach.NewHandler(coachSvc, coachRepo))
	movement.RegisterRoutes(protected, movement.NewHandler(movement.NewService(movement.NewRepository(db))))
	plan.RegisterRoutes(protected, plan.NewHandler(plan.NewService(plan.NewRepository(db), athleteSvc)))

	return router, db
}

type movementResponse struct {
	ID          string  `json:"id"`
	CoachID     *string `json:"coach_id"`
	Name        string  `json:"name"`
	Category    *string `json:"category"`
	Description *string `json:"description"`
	MuscleGroup *string `json:"muscle_group"`
	Equipment   *string `json:"equipment"`
	MediaURL    *string `json:"media_url"`
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

func requireErrorCode(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, wantCode string) {
	t.Helper()

	env := decodeEnvelope(t, rec, wantStatus)
	assert.False(t, env.Success)
	require.NotNil(t, env.Error)
	assert.Equal(t, wantCode, env.Error.Code)
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

type testCoach struct {
	id    string
	token string
}

func signupCoach(t *testing.T, router *gin.Engine, phone string) testCoach {
	t.Helper()

	credentials := fmt.Sprintf(`{"first_name":"Test","last_name":"Coach","phone":%q,"password":"correct-horse-battery-staple"}`, phone)

	var signedUp struct {
		ID string `json:"id"`
	}
	env := decodeEnvelope(t, do(t, router, http.MethodPost, "/api/v1/auth/signup", "", credentials), http.StatusCreated)
	require.NoError(t, json.Unmarshal(env.Data, &signedUp))

	var loggedIn struct {
		Token string `json:"token"`
	}
	env = decodeEnvelope(t, do(t, router, http.MethodPost, "/api/v1/auth/login", "", credentials), http.StatusOK)
	require.NoError(t, json.Unmarshal(env.Data, &loggedIn))
	require.NotEmpty(t, loggedIn.Token)

	return testCoach{id: signedUp.ID, token: loggedIn.Token}
}

func listMovements(t *testing.T, router *gin.Engine, token, query string) []movementResponse {
	t.Helper()

	env := decodeEnvelope(t, do(t, router, http.MethodGet, movementsPath+query, token, ""), http.StatusOK)
	assert.True(t, env.Success)

	var out []movementResponse
	require.NoError(t, json.Unmarshal(env.Data, &out))

	return out
}

func createMovement(t *testing.T, router *gin.Engine, token, body string) movementResponse {
	t.Helper()

	env := decodeEnvelope(t, do(t, router, http.MethodPost, movementsPath, token, body), http.StatusCreated)

	var out movementResponse
	require.NoError(t, json.Unmarshal(env.Data, &out))

	return out
}

func names(movements []movementResponse) []string {
	out := make([]string, 0, len(movements))
	for _, m := range movements {
		out = append(out, m.Name)
	}
	return out
}

func findMovement(t *testing.T, db *gorm.DB, id string) movement.Movement {
	t.Helper()

	var m movement.Movement
	require.NoError(t, db.Unscoped().First(&m, "id = ?", id).Error)

	return m
}

func TestList_UniversalMovementsInEveryCoachsList(t *testing.T) {
	router, _ := newTestRouter(t)

	coachA := signupCoach(t, router, "09121000001")
	coachB := signupCoach(t, router, "09121000002")

	for name, c := range map[string]testCoach{"coach A": coachA, "coach B": coachB} {
		t.Run(name+" starts with the seeded library", func(t *testing.T) {
			listed := listMovements(t, router, c.token, "")

			require.Len(t, listed, seededCount)
			for _, m := range listed {
				assert.Nil(t, m.CoachID, "%s is universal", m.Name)
			}
			assert.Contains(t, names(listed), seededBackSquatName)
		})
	}

	own := createMovement(t, router, coachA.token, `{"name":"Goblet squat","muscle_group":"legs","equipment":"kettlebell"}`)

	t.Run("a coach's own movement joins their list", func(t *testing.T) {
		listed := listMovements(t, router, coachA.token, "")

		assert.Len(t, listed, seededCount+1)
		assert.Contains(t, names(listed), own.Name)
	})

	t.Run("but not another coach's", func(t *testing.T) {
		listed := listMovements(t, router, coachB.token, "")

		assert.Len(t, listed, seededCount)
		assert.NotContains(t, names(listed), own.Name)
	})
}

func TestList_SeededLibraryFillsEveryFilter(t *testing.T) {
	router, db := newTestRouter(t)
	c := signupCoach(t, router, "09121000003")

	var incomplete int64
	require.NoError(t, db.Model(&movement.Movement{}).
		Where("coach_id IS NULL AND (muscle_group IS NULL OR equipment IS NULL OR category IS NULL OR media_url IS NOT NULL)").
		Count(&incomplete).Error)
	assert.Zero(t, incomplete, "seeded rows have muscle group, equipment and category, and no media yet")

	filters := map[string][]string{
		"muscle_group": {"chest", "back", "shoulders", "arms", "legs", "core", "full_body"},
		"equipment":    {"bodyweight", "barbell", "dumbbell", "kettlebell", "machine", "cable", "band", "other"},
	}
	for param, values := range filters {
		for _, value := range values {
			t.Run(param+"="+value, func(t *testing.T) {
				listed := listMovements(t, router, c.token, "?"+param+"="+value)

				require.NotEmpty(t, listed, "a fresh account's filter must not come back empty")
				for _, m := range listed {
					got := m.MuscleGroup
					if param == "equipment" {
						got = m.Equipment
					}
					require.NotNil(t, got)
					assert.Equal(t, value, *got, m.Name)
				}
			})
		}
	}
}

func TestList_FiltersAndSearch(t *testing.T) {
	router, _ := newTestRouter(t)

	c := signupCoach(t, router, "09121000004")
	createMovement(t, router, c.token, `{"name":"Goblet squat","muscle_group":"legs","equipment":"kettlebell"}`)
	createMovement(t, router, c.token, `{"name":"100% effort sprint","category":"cardio"}`)

	for name, tc := range map[string]struct {
		query string
		want  []string
	}{
		"muscle group and equipment combine with AND": {
			query: "?muscle_group=legs&equipment=barbell",
			want:  []string{seededBackSquatName, "Romanian deadlift"},
		},
		"a coach's own movement is filtered like a seeded one": {
			query: "?equipment=kettlebell",
			want:  []string{"Goblet squat", "Kettlebell swing"},
		},
		"search is case-insensitive and partial, across own and universal": {
			query: "?q=SQUAT",
			want:  []string{seededBackSquatName, "Bodyweight squat", "Goblet squat"},
		},
		"search combines with filters": {
			query: "?q=squat&equipment=bodyweight",
			want:  []string{"Bodyweight squat"},
		},
		"a percent sign matches literally": {
			query: "?q=0%25",
			want:  []string{"100% effort sprint"},
		},
		"filter values are matched case-insensitively": {
			query: "?muscle_group=CORE",
			want:  []string{"Plank"},
		},
		"no match is an empty list": {
			query: "?q=zercher",
			want:  []string{},
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, names(listMovements(t, router, c.token, tc.query)), "ordered by name")
		})
	}

	t.Run("an unknown filter value is a 400, not an empty list", func(t *testing.T) {
		requireErrorCode(t, do(t, router, http.MethodGet, movementsPath+"?muscle_group=quads", c.token, ""),
			http.StatusBadRequest, httpx.CodeValidationFailed)
	})
}

func TestCreate(t *testing.T) {
	router, db := newTestRouter(t)

	c := signupCoach(t, router, "09121000005")
	other := signupCoach(t, router, "09121000006")

	t.Run("owner comes from the token, never the body", func(t *testing.T) {
		body := fmt.Sprintf(`{"name":"Bulgarian split squat","category":"strength","muscle_group":"Legs","equipment":"dumbbell","coach_id":%q}`, other.id)
		created := createMovement(t, router, c.token, body)

		require.NotNil(t, created.CoachID)
		assert.Equal(t, c.id, *created.CoachID)
		assert.Equal(t, legs, *created.MuscleGroup, "stored lowercase")
		assert.Nil(t, created.MediaURL)

		stored := findMovement(t, db, created.ID)
		require.NotNil(t, stored.CoachID)
		assert.Equal(t, c.id, stored.CoachID.String())
	})

	t.Run("bad input reports every field", func(t *testing.T) {
		env := decodeEnvelope(t, do(t, router, http.MethodPost, movementsPath, c.token, `{"muscle_group":"quads","equipment":"sled"}`),
			http.StatusBadRequest)

		require.NotNil(t, env.Error)
		assert.Equal(t, httpx.CodeValidationFailed, env.Error.Code)
		assert.Equal(t, []string{"equipment", "muscle_group", "name"}, slices.Sorted(maps.Keys(env.Error.Fields)))
	})
}

func TestUpdateAndDelete_OwnMovement(t *testing.T) {
	router, db := newTestRouter(t)
	c := signupCoach(t, router, "09121000007")

	created := createMovement(t, router, c.token, `{"name":"Goblet squat","category":"strength","description":"Hold at the chest","muscle_group":"legs","equipment":"kettlebell"}`)
	path := movementsPath + "/" + created.ID

	mediaURL := "https://example.com/goblet.mp4"
	require.NoError(t, db.Model(&movement.Movement{}).Where("id = ?", created.ID).Update("media_url", mediaURL).Error)

	t.Run("update replaces every editable field", func(t *testing.T) {
		env := decodeEnvelope(t, do(t, router, http.MethodPut, path, c.token, `{"name":"Goblet box squat","muscle_group":"legs","equipment":"dumbbell"}`),
			http.StatusOK)

		var updated movementResponse
		require.NoError(t, json.Unmarshal(env.Data, &updated))
		assert.Equal(t, "Goblet box squat", updated.Name)
		assert.Equal(t, "dumbbell", *updated.Equipment)
		assert.Nil(t, updated.Category, "an omitted optional field is cleared")
		assert.Nil(t, updated.Description)
		require.NotNil(t, updated.MediaURL, "media_url isn't editable here, so it's kept")
		assert.Equal(t, mediaURL, *updated.MediaURL)

		stored := findMovement(t, db, created.ID)
		assert.Equal(t, "Goblet box squat", stored.Name)
		assert.Nil(t, stored.Category)
		assert.True(t, stored.UpdatedAt.After(stored.CreatedAt), "updated_at moves on an edit")
	})

	t.Run("update validates like create", func(t *testing.T) {
		env := decodeEnvelope(t, do(t, router, http.MethodPut, path, c.token, `{"name":" "}`), http.StatusBadRequest)
		require.NotNil(t, env.Error)
		assert.Equal(t, httpx.MsgFieldRequired, env.Error.Fields["name"])
		assert.Equal(t, "Goblet box squat", findMovement(t, db, created.ID).Name)
	})

	t.Run("delete removes it from the library, softly", func(t *testing.T) {
		env := decodeEnvelope(t, do(t, router, http.MethodDelete, path, c.token, ""), http.StatusOK)
		assert.True(t, env.Success)
		assert.Equal(t, "null", string(env.Data))

		assert.NotContains(t, names(listMovements(t, router, c.token, "")), "Goblet box squat")
		assert.True(t, findMovement(t, db, created.ID).DeletedAt.Valid, "the row stays, marked deleted")
	})

	t.Run("a deleted movement is gone for good", func(t *testing.T) {
		requireErrorCode(t, do(t, router, http.MethodPut, path, c.token, `{"name":"Back again"}`), http.StatusNotFound, httpx.CodeNotFound)
		requireErrorCode(t, do(t, router, http.MethodDelete, path, c.token, ""), http.StatusNotFound, httpx.CodeNotFound)
	})
}

// Universal movements are visible, so a 403 says nothing new. Another coach's movement is not, so it
// gets the same 404 as one that doesn't exist, like every other tenant-owned resource.
func TestUpdateAndDelete_OnlyTheOwnerMayChangeAMovement(t *testing.T) {
	router, db := newTestRouter(t)

	c := signupCoach(t, router, "09121000008")
	other := signupCoach(t, router, "09121000009")
	othersLunge := createMovement(t, router, other.token, `{"name":"Other coach's lunge"}`)

	for name, tc := range map[string]struct {
		id         string
		wantName   string
		wantStatus int
		wantCode   string
	}{
		"universal movement":       {id: seededBackSquat, wantName: seededBackSquatName, wantStatus: http.StatusForbidden, wantCode: httpx.CodeForbidden},
		"another coach's movement": {id: othersLunge.ID, wantName: othersLunge.Name, wantStatus: http.StatusNotFound, wantCode: httpx.CodeNotFound},
	} {
		t.Run(name, func(t *testing.T) {
			path := movementsPath + "/" + tc.id

			requireErrorCode(t, do(t, router, http.MethodPut, path, c.token, `{"name":"Squat"}`), tc.wantStatus, tc.wantCode)
			requireErrorCode(t, do(t, router, http.MethodPut, path, c.token, `{}`), tc.wantStatus, tc.wantCode)
			requireErrorCode(t, do(t, router, http.MethodDelete, path, c.token, ""), tc.wantStatus, tc.wantCode)

			stored := findMovement(t, db, tc.id)
			assert.Equal(t, tc.wantName, stored.Name, "not renamed")
			assert.False(t, stored.DeletedAt.Valid, "not deleted")
		})
	}

	t.Run("the owner can still change their own", func(t *testing.T) {
		decodeEnvelope(t, do(t, router, http.MethodPut, movementsPath+"/"+othersLunge.ID, other.token, `{"name":"Reverse lunge"}`), http.StatusOK)
	})
}

func TestDelete_MovementUsedInAPlan(t *testing.T) {
	router, db := newTestRouter(t)
	c := signupCoach(t, router, "09121000010")

	used := createMovement(t, router, c.token, `{"name":"Goblet squat","category":"strength"}`)
	planID := insertPlanUsing(t, db, uuid.MustParse(c.id), uuid.MustParse(used.ID))

	decodeEnvelope(t, do(t, router, http.MethodDelete, movementsPath+"/"+used.ID, c.token, ""), http.StatusOK)

	var refs int64
	require.NoError(t, db.Model(&plan.BlockMovement{}).Where("movement_id = ?", used.ID).Count(&refs).Error)
	assert.Equal(t, int64(1), refs, "the plan's reference survives the delete")

	env := decodeEnvelope(t, do(t, router, http.MethodGet, "/api/v1/plans/"+planID.String(), c.token, ""), http.StatusOK)
	var detail struct {
		Days []struct {
			Blocks []struct {
				Movements []struct {
					MovementID string `json:"movement_id"`
					Name       string `json:"name"`
				} `json:"movements"`
			} `json:"blocks"`
		} `json:"days"`
	}
	require.NoError(t, json.Unmarshal(env.Data, &detail))
	require.Len(t, detail.Days, 1)
	require.Len(t, detail.Days[0].Blocks, 1)
	require.Len(t, detail.Days[0].Blocks[0].Movements, 1)
	assert.Equal(t, used.Name, detail.Days[0].Blocks[0].Movements[0].Name, "the plan still names the deleted movement")
}

// insertPlanUsing writes the plan tree straight to the tables; building it is the plan module's concern.
func insertPlanUsing(t *testing.T, db *gorm.DB, coachID, movementID uuid.UUID) uuid.UUID {
	t.Helper()

	a := athlete.Athlete{ID: uuid.New(), CoachID: coachID, FirstName: "Sara", LastName: "Ahmadi", Phone: "09372144430"}
	p := plan.Plan{ID: uuid.New(), AthleteID: a.ID, Title: "Base", StartDate: time.Now().UTC().Truncate(24 * time.Hour)}
	d := plan.Day{ID: uuid.New(), PlanID: p.ID, Label: "A", OrderIndex: 0}
	b := plan.Block{ID: uuid.New(), DayID: d.ID, OrderIndex: 0, Sets: 3}
	reps := 10
	bm := plan.BlockMovement{ID: uuid.New(), BlockID: b.ID, MovementID: movementID, Reps: &reps, OrderInBlock: 0}

	for _, row := range []any{&a, &p, &d, &b, &bm} {
		require.NoError(t, db.Omit(clause.Associations).Create(row).Error)
	}

	return p.ID
}
