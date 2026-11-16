package postgres_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	adapter "github.com/puriice/waypoint/authentication/internal/adapter/repository"
	"github.com/puriice/waypoint/authentication/internal/adapter/repository/postgres"
	"github.com/puriice/waypoint/authentication/internal/core/port/repository"
)

// databaseURLEnv names the connection string of a PostgreSQL 18+ server
// (the migrations use uuidv7()). The test creates its own schema and drops it
// afterwards, so it never touches existing tables.
const databaseURLEnv = "AUTHENTICATION_TEST_DATABASE_URL"

const migrationsDir = "../../../../migrations"

func TestAuthenticationRepository(t *testing.T) {
	pool := newTestPool(t)

	adapter.RunAuthenticationSuite(t, func(t *testing.T) repository.Authentication {
		return postgres.NewAuthentication(pool, time.Hour)
	})
}

// newTestPool connects to the test database, creates an isolated schema,
// applies every migration's Up section to it, and returns a pool whose
// connections use that schema. The schema is dropped when the test ends.
func newTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv(databaseURLEnv)
	if dsn == "" {
		t.Skipf("%s is not set; skipping PostgreSQL repository tests", databaseURLEnv)
	}

	ctx := context.Background()
	schema := "auth_test_" + strings.ReplaceAll(uuid.NewV4().String(), "-", "")

	admin, err := pgx.Connect(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(func() { admin.Close(context.Background()) })

	_, err = admin.Exec(ctx, "CREATE SCHEMA "+pgx.Identifier{schema}.Sanitize())
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := admin.Exec(context.Background(), "DROP SCHEMA "+pgx.Identifier{schema}.Sanitize()+" CASCADE")
		if err != nil {
			t.Errorf("drop test schema %s: %v", schema, err)
		}
	})

	cfg, err := pgxpool.ParseConfig(dsn)
	require.NoError(t, err)
	cfg.ConnConfig.RuntimeParams["search_path"] = schema

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	applyMigrations(t, pool)

	return pool
}

func applyMigrations(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	require.NoError(t, err)
	require.NotEmpty(t, files, "no migrations found in %s", migrationsDir)
	sort.Strings(files)

	for _, file := range files {
		content, err := os.ReadFile(file)
		require.NoError(t, err)

		up, err := upSection(string(content))
		require.NoError(t, err, file)

		_, err = pool.Exec(context.Background(), up)
		require.NoError(t, err, "apply %s", filepath.Base(file))
	}
}

// upSection returns the SQL between the "-- +godwit Up" and
// "-- +godwit Down" markers of a godwit migration.
func upSection(migration string) (string, error) {
	const upMarker, downMarker = "-- +godwit Up", "-- +godwit Down"

	start := strings.Index(migration, upMarker)
	if start == -1 {
		return "", fmt.Errorf("missing %q marker", upMarker)
	}
	up := migration[start+len(upMarker):]

	if end := strings.Index(up, downMarker); end != -1 {
		up = up[:end]
	}

	return up, nil
}
