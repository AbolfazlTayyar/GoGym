package testutil_test

import (
	"testing"

	"github.com/AbolfazlTayyar/gogym/internal/models"
	"github.com/AbolfazlTayyar/gogym/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestNewDB_MigratedSchema proves the harness itself works: a container
// comes up, migrations apply, and a *gorm.DB against it can round-trip a
// row through the migrated schema. It is not feature coverage.
func TestNewDB_MigratedSchema(t *testing.T) {
	db := testutil.NewDB(t)

	coach := models.Coach{
		ID:           uuid.New(),
		FirstName:    "Ada",
		LastName:     "Lovelace",
		Phone:        "+15550000000",
		PasswordHash: "hashed",
	}
	require.NoError(t, db.Create(&coach).Error)

	var got models.Coach
	require.NoError(t, db.First(&got, "id = ?", coach.ID).Error)
	require.Equal(t, coach.Phone, got.Phone)
}
