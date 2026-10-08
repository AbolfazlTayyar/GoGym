package tenant_test

import (
	"testing"

	"github.com/AbolfazlTayyar/gogym/internal/athlete"
	coachpkg "github.com/AbolfazlTayyar/gogym/internal/coach"
	"github.com/AbolfazlTayyar/gogym/internal/movement"
	"github.com/AbolfazlTayyar/gogym/internal/tenant"
	"github.com/AbolfazlTayyar/gogym/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScope_FiltersByCoach(t *testing.T) {
	db := testutil.NewDB(t)

	coachA := coachpkg.Coach{ID: uuid.New(), FirstName: "A", LastName: "Coach", Phone: "09121000001", PasswordHash: "x"}
	coachB := coachpkg.Coach{ID: uuid.New(), FirstName: "B", LastName: "Coach", Phone: "09121000002", PasswordHash: "x"}
	require.NoError(t, db.Create(&coachA).Error)
	require.NoError(t, db.Create(&coachB).Error)

	athleteA := athlete.Athlete{ID: uuid.New(), CoachID: coachA.ID, FirstName: "Ann", LastName: "Athlete", Phone: "09121000003"}
	athleteB := athlete.Athlete{ID: uuid.New(), CoachID: coachB.ID, FirstName: "Bob", LastName: "Athlete", Phone: "09121000004"}
	require.NoError(t, db.Create(&athleteA).Error)
	require.NoError(t, db.Create(&athleteB).Error)

	var asCoachA []athlete.Athlete
	require.NoError(t, tenant.Scope(db, coachA.ID).Find(&asCoachA).Error)
	require.Len(t, asCoachA, 1)
	require.Equal(t, athleteA.ID, asCoachA[0].ID)

	var asCoachB []athlete.Athlete
	require.NoError(t, tenant.Scope(db, coachB.ID).Find(&asCoachB).Error)
	require.Len(t, asCoachB, 1)
	require.Equal(t, athleteB.ID, asCoachB[0].ID)

	var lookup athlete.Athlete
	err := tenant.Scope(db, coachA.ID).First(&lookup, "id = ?", athleteB.ID).Error
	require.Error(t, err)
}

func TestScopeWithUniversal_AddsUniversalRowsOnly(t *testing.T) {
	db := testutil.NewDB(t)

	coachA := coachpkg.Coach{ID: uuid.New(), FirstName: "A", LastName: "Coach", Phone: "09121000011", PasswordHash: "x"}
	coachB := coachpkg.Coach{ID: uuid.New(), FirstName: "B", LastName: "Coach", Phone: "09121000012", PasswordHash: "x"}
	require.NoError(t, db.Create(&coachA).Error)
	require.NoError(t, db.Create(&coachB).Error)

	var created []uuid.UUID
	insert := func(coachID *uuid.UUID, name string) movement.Movement {
		m := movement.Movement{ID: uuid.New(), CoachID: coachID, Name: name}
		require.NoError(t, db.Create(&m).Error)
		created = append(created, m.ID)
		return m
	}
	universal := insert(nil, "Squat")
	ownA := insert(&coachA.ID, "A's lunge")
	insert(&coachB.ID, "B's lunge")
	retired := insert(nil, "Retired")
	require.NoError(t, db.Delete(&retired).Error)

	// Limited to this test's rows, since the migrations also seed universal movements.
	var visible []movement.Movement
	require.NoError(t, tenant.ScopeWithUniversal(db, coachA.ID).Where("id IN ?", created).Order("name").Find(&visible).Error)

	names := make([]string, 0, len(visible))
	for _, m := range visible {
		names = append(names, m.Name)
	}
	assert.Equal(t, []string{ownA.Name, universal.Name}, names, "own and universal, not another coach's or soft-deleted")

	var leaked []movement.Movement
	require.NoError(t, tenant.ScopeWithUniversal(db, coachA.ID).Where("name = ?", "B's lunge").Find(&leaked).Error)
	assert.Empty(t, leaked, "the OR stays grouped, so a chained condition can't widen it to another coach's rows")
}
