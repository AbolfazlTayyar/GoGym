package plan_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/AbolfazlTayyar/gogym/internal/plan"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const (
	shareCoachPhone   = "09120000001"
	shareAthletePhone = "09350000002"
)

type shareLink struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

type sharedPlan struct {
	Status    string `json:"status"`
	ExpiresAt string `json:"expires_at"`
	Coach     struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Phone     string `json:"phone"`
	} `json:"coach"`
	Athlete *struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	} `json:"athlete"`
	Plan *struct {
		Title     string `json:"title"`
		StartDate string `json:"start_date"`
		Days      []struct {
			Label  string `json:"label"`
			Blocks []struct {
				Movements []struct {
					Name string `json:"name"`
				} `json:"movements"`
			} `json:"blocks"`
		} `json:"days"`
	} `json:"plan"`
}

func sharePath(planID uuid.UUID) string {
	return "/api/v1/plans/" + planID.String() + "/share"
}

func sharedPath(token string) string {
	return "/api/v1/public/plans/" + token
}

func share(t *testing.T, router *gin.Engine, token string, planID uuid.UUID) shareLink {
	t.Helper()

	env := decodeEnvelope(t, do(t, router, http.MethodPost, sharePath(planID), token, ""), http.StatusCreated)
	var link shareLink
	require.NoError(t, json.Unmarshal(env.Data, &link))
	return link
}

// getShared sends no token: the share token in the path is the only access.
func getShared(t *testing.T, router *gin.Engine, shareToken string, wantStatus int) (sharedPlan, envelope) {
	t.Helper()

	env := decodeEnvelope(t, do(t, router, http.MethodGet, sharedPath(shareToken), "", ""), wantStatus)
	var out sharedPlan
	if wantStatus == http.StatusOK {
		require.NoError(t, json.Unmarshal(env.Data, &out))
	}
	return out, env
}

type shareFixture struct {
	router    *gin.Engine
	db        *gorm.DB
	token     string
	athleteID string
	planID    uuid.UUID
}

// newShareFixture is a one-day plan with one line, so the shared tree has something in every level.
func newShareFixture(t *testing.T) shareFixture {
	t.Helper()

	router, db := newTestRouter(t)
	token := createCoach(t, router, db, shareCoachPhone)
	athleteID := createAthlete(t, router, token, shareAthletePhone)
	planID := insertPlan(t, db, athleteID, "برنامه آبان", daysFromToday(-1))

	dayID, err := insertDay(db, planID, "A", 0)
	require.NoError(t, err)
	blockID := insertBlock(t, db, dayID, 0, 3, nil)
	insertBlockMovement(t, db, blockID, insertMovement(t, db, "Squat", "strength"), 10, 0)

	return shareFixture{router: router, db: db, token: token, athleteID: athleteID, planID: planID}
}

func TestSharePlan_OpensThePlanWithoutLogin(t *testing.T) {
	f := newShareFixture(t)

	before := time.Now()
	link := share(t, f.router, f.token, f.planID)
	assert.Len(t, link.Token, 22, "128 bits in unpadded base64url")

	expiresAt, err := time.Parse(time.RFC3339, link.ExpiresAt)
	require.NoError(t, err)
	assert.WithinDuration(t, before.Add(30*24*time.Hour), expiresAt, time.Minute)

	rec := do(t, f.router, http.MethodGet, sharedPath(link.Token), "", "")
	assert.Equal(t, "noindex", rec.Header().Get("X-Robots-Tag"))
	env := decodeEnvelope(t, rec, http.StatusOK)

	var got sharedPlan
	require.NoError(t, json.Unmarshal(env.Data, &got))
	assert.Equal(t, "active", got.Status)
	assert.Equal(t, link.ExpiresAt, got.ExpiresAt)
	assert.Equal(t, "Test", got.Coach.FirstName)
	assert.Equal(t, "Coach", got.Coach.LastName)
	assert.Equal(t, shareCoachPhone, got.Coach.Phone)
	require.NotNil(t, got.Athlete)
	assert.Equal(t, "Sara", got.Athlete.FirstName)
	assert.Equal(t, "Ahmadi", got.Athlete.LastName)
	require.NotNil(t, got.Plan)
	assert.Equal(t, "برنامه آبان", got.Plan.Title)
	require.Len(t, got.Plan.Days, 1)
	assert.Equal(t, "A", got.Plan.Days[0].Label)
	require.Len(t, got.Plan.Days[0].Blocks, 1)
	assert.Equal(t, "Squat", got.Plan.Days[0].Blocks[0].Movements[0].Name)

	raw := string(env.Data)
	for _, leak := range []string{shareAthletePhone, f.athleteID, f.planID.String(), "source_plan", "share_token"} {
		assert.NotContains(t, raw, leak)
	}
}

func TestSharePlan_DoesNotTouchUpdatedAt(t *testing.T) {
	f := newShareFixture(t)

	var before plan.Plan
	require.NoError(t, f.db.First(&before, "id = ?", f.planID).Error)

	share(t, f.router, f.token, f.planID)

	var after plan.Plan
	require.NoError(t, f.db.First(&after, "id = ?", f.planID).Error)
	assert.True(t, before.UpdatedAt.Equal(after.UpdatedAt))
}

func TestSharePlan_AgainReplacesTheOldLink(t *testing.T) {
	f := newShareFixture(t)

	first := share(t, f.router, f.token, f.planID)
	second := share(t, f.router, f.token, f.planID)
	require.NotEqual(t, first.Token, second.Token)

	_, env := getShared(t, f.router, first.Token, http.StatusNotFound)
	assert.Equal(t, "not_found", env.Error.Code)
	got, _ := getShared(t, f.router, second.Token, http.StatusOK)
	assert.Equal(t, "active", got.Status)
}

func TestSharePlan_AnotherCoachsPlanIsNotFound(t *testing.T) {
	f := newShareFixture(t)
	otherToken := createCoach(t, f.router, f.db, "09120000003")

	decodeEnvelope(t, do(t, f.router, http.MethodPost, sharePath(f.planID), otherToken, ""), http.StatusNotFound)
	decodeEnvelope(t, do(t, f.router, http.MethodPost, sharePath(f.planID), "", ""), http.StatusUnauthorized)

	var p plan.Plan
	require.NoError(t, f.db.First(&p, "id = ?", f.planID).Error)
	assert.Nil(t, p.ShareToken, "a refused share must not create a link")
}

func TestSharedPlan_ExpiredStillNamesTheCoach(t *testing.T) {
	f := newShareFixture(t)
	link := share(t, f.router, f.token, f.planID)
	require.NoError(t, f.db.Model(&plan.Plan{}).Where("id = ?", f.planID).
		Update("share_expires_at", time.Now().Add(-time.Minute)).Error)

	got, env := getShared(t, f.router, link.Token, http.StatusOK)
	assert.Equal(t, "expired", got.Status)
	assert.Equal(t, shareCoachPhone, got.Coach.Phone)
	assert.Nil(t, got.Athlete)
	assert.Nil(t, got.Plan)

	raw := string(env.Data)
	assert.Contains(t, raw, `"athlete":null`)
	assert.Contains(t, raw, `"plan":null`)
	assert.False(t, strings.Contains(raw, "برنامه آبان"), "an expired link must not show the plan")
}

func TestSharedPlan_UnknownDeletedPlanOrAthleteIsNotFound(t *testing.T) {
	t.Run("unknown token", func(t *testing.T) {
		f := newShareFixture(t)
		getShared(t, f.router, "nope", http.StatusNotFound)
	})

	t.Run("deleted plan", func(t *testing.T) {
		f := newShareFixture(t)
		link := share(t, f.router, f.token, f.planID)
		decodeEnvelope(t, do(t, f.router, http.MethodDelete, planPath(f.planID.String()), f.token, ""), http.StatusOK)
		getShared(t, f.router, link.Token, http.StatusNotFound)
	})

	t.Run("soft-deleted athlete", func(t *testing.T) {
		f := newShareFixture(t)
		link := share(t, f.router, f.token, f.planID)
		decodeEnvelope(t, do(t, f.router, http.MethodDelete, "/api/v1/athletes/"+f.athleteID, f.token, ""), http.StatusOK)
		getShared(t, f.router, link.Token, http.StatusNotFound)
	})
}

func TestSharedPlan_IsRateLimitedPerIP(t *testing.T) {
	f := newShareFixture(t)
	link := share(t, f.router, f.token, f.planID)

	for i := range 30 {
		require.Equal(t, http.StatusOK, do(t, f.router, http.MethodGet, sharedPath(link.Token), "", "").Code, "request %d", i+1)
	}
	_, env := getShared(t, f.router, link.Token, http.StatusTooManyRequests)
	assert.Equal(t, "rate_limited", env.Error.Code)
}
