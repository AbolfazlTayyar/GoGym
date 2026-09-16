// Package testutil provides a throwaway Postgres instance, migrated with
// the same migrations/*.sql files used in production, for integration
// tests in other modules to run against.
package testutil

import (
	"context"
	"errors"
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

// NewDB starts a throwaway Postgres container, applies every migration in
// migrations/ against it, and returns a *gorm.DB connected to it. The
// container is terminated automatically via t.Cleanup.
//
// It skips the test when run with `go test -short` (see `make test-unit`),
// so integration tests only run as part of `make test`.
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

	gormDB, err := gorm.Open(pgdriver.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	return gormDB
}

// applyMigrations runs every up migration in migrations/ against dsn.
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

// migrationsSourceURL resolves the migrations/ directory to a file:// URL
// golang-migrate can use, regardless of which package's test invokes NewDB
// or which OS it runs on. filepath.ToSlash is enough on both: on Unix the
// path already starts with "/", giving "file:///home/..."; on Windows it
// starts with a drive letter, giving "file://C:/...", which golang-migrate's
// file source parses as host "C:" + path "/..." — a valid Windows path.
func migrationsSourceURL() string {
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(file), "..", "..", "migrations")

	return "file://" + filepath.ToSlash(dir)
}
