package testutil_test

import (
	"testing"

	coachpkg "github.com/AbolfazlTayyar/gogym/internal/coach"
	"github.com/AbolfazlTayyar/gogym/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestNewDB_MigratedSchema checks the harness itself, not feature behavior.
func TestNewDB_MigratedSchema(t *testing.T) {
	db := testutil.NewDB(t)

	c := coachpkg.Coach{
		ID:           uuid.New(),
		FirstName:    "Ada",
		LastName:     "Lovelace",
		Phone:        "09121234567",
		PasswordHash: "hashed",
	}
	require.NoError(t, db.Create(&c).Error)

	var got coachpkg.Coach
	require.NoError(t, db.First(&got, "id = ?", c.ID).Error)
	require.Equal(t, c.Phone, got.Phone)
}
