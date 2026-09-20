package tenant_test

import (
	"testing"

	coachpkg "github.com/AbolfazlTayyar/gogym/internal/coach"
	"github.com/AbolfazlTayyar/gogym/internal/models"
	"github.com/AbolfazlTayyar/gogym/internal/tenant"
	"github.com/AbolfazlTayyar/gogym/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestScope_FiltersByCoach proves the scoping helper actually filters:
// querying athlete through tenant.Scope as coach A never returns coach B's
// rows, and vice versa.
func TestScope_FiltersByCoach(t *testing.T) {
	db := testutil.NewDB(t)

	coachA := coachpkg.Coach{ID: uuid.New(), FirstName: "A", LastName: "Coach", Phone: "09121000001", PasswordHash: "x"}
	coachB := coachpkg.Coach{ID: uuid.New(), FirstName: "B", LastName: "Coach", Phone: "09121000002", PasswordHash: "x"}
	require.NoError(t, db.Create(&coachA).Error)
	require.NoError(t, db.Create(&coachB).Error)

	athleteA := models.Athlete{ID: uuid.New(), CoachID: coachA.ID, FirstName: "Ann", LastName: "Athlete", Phone: "09121000003"}
	athleteB := models.Athlete{ID: uuid.New(), CoachID: coachB.ID, FirstName: "Bob", LastName: "Athlete", Phone: "09121000004"}
	require.NoError(t, db.Create(&athleteA).Error)
	require.NoError(t, db.Create(&athleteB).Error)

	var asCoachA []models.Athlete
	require.NoError(t, tenant.Scope(db, coachA.ID).Find(&asCoachA).Error)
	require.Len(t, asCoachA, 1)
	require.Equal(t, athleteA.ID, asCoachA[0].ID)

	var asCoachB []models.Athlete
	require.NoError(t, tenant.Scope(db, coachB.ID).Find(&asCoachB).Error)
	require.Len(t, asCoachB, 1)
	require.Equal(t, athleteB.ID, asCoachB[0].ID)

	// coach A's scoped query must not be able to reach athlete B's row by id.
	var lookup models.Athlete
	err := tenant.Scope(db, coachA.ID).First(&lookup, "id = ?", athleteB.ID).Error
	require.Error(t, err)
}
