// Package testutil provides a migrated throwaway Postgres for integration tests.
package testutil

import (
	"context"
	"errors"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	// Registers the postgres database driver and file source with golang-migrate.
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/require"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	postgresImage = "postgres:16"
	testDBName    = "gogym_test"
	testDBUser    = "gogym"
	testDBPass    = "gogym"
)

// NewDB skips the test under go test -short, which is how unit-only runs exclude integration tests.
func NewDB(t *testing.T) *gorm.DB {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()

	container, err := tcpostgres.Run(ctx, postgresImage,
		tcpostgres.WithDatabase(testDBName),
		tcpostgres.WithUsername(testDBUser),
		tcpostgres.WithPassword(testDBPass),
		tcpostgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, container.Terminate(context.Background()))
	})

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	applyMigrations(t, dsn)

	gormDB, err := gorm.Open(pgdriver.Open(dsn), &gorm.Config{TranslateError: true})
	require.NoError(t, err)

	return gormDB
}

func applyMigrations(t *testing.T, dsn string) {
	t.Helper()

	m, err := migrate.New(migrationsSourceURL(), dsn)
	require.NoError(t, err)
	t.Cleanup(func() {
		srcErr, dbErr := m.Close()
		require.NoError(t, srcErr)
		require.NoError(t, dbErr)
	})

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		require.NoError(t, err)
	}
}

// ToSlash suffices on Windows too: golang-migrate parses "file://C:/..." as host "C:" plus a valid path.
func migrationsSourceURL() string {
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(file), "..", "..", "..", "migrations")

	return "file://" + filepath.ToSlash(dir)
}

// Cookie returns the value of the named cookie the response set, failing the test if it set none.
func Cookie(t *testing.T, rec *httptest.ResponseRecorder, name string) string {
	t.Helper()

	for _, c := range rec.Result().Cookies() {
		if c.Name == name {
			require.NotEmpty(t, c.Value, "cookie %s is empty", name)
			return c.Value
		}
	}
	require.Failf(t, "cookie not set", "the response set no %s cookie", name)
	return ""
}
