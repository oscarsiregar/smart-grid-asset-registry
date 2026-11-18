// Package pgtest gives integration tests a real PostgreSQL database
// (constitution, Template Exception 11). Tests that call Open are skipped
// unless TEST_DATABASE_URL points at a disposable database, for example the
// compose "asset_registry_test" database.
package pgtest

import (
	"context"
	"os"
	"testing"

	"smart-grid-asset-registry/migrations"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// EnvDatabaseURL names the variable holding the test database URL.
const EnvDatabaseURL = "TEST_DATABASE_URL"

// lockKey identifies the session advisory lock that serialises tests using
// the shared test database, also across test packages, which `go test ./...`
// runs in parallel processes.
const lockKey int64 = 7_334_455_001

// Open skips the test when EnvDatabaseURL is unset. Otherwise it waits for
// exclusive use of the test database, applies the embedded migrations,
// empties the import tables and returns a connection. When the test ends the
// tables are emptied again and the database is released.
//
// Within a package, tests that call Open must not use t.Parallel.
func Open(t testing.TB) *gorm.DB {
	t.Helper()

	databaseURL := os.Getenv(EnvDatabaseURL)
	if databaseURL == "" {
		t.Skipf("%s is not set; skipping real-PostgreSQL test", EnvDatabaseURL)
	}

	if err := migrations.Up(databaseURL); err != nil {
		t.Fatalf("pgtest: %v", err)
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger:                 gormlogger.Default.LogMode(gormlogger.Silent),
		SkipDefaultTransaction: true,
	})
	if err != nil {
		t.Fatalf("pgtest: open database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("pgtest: %v", err)
	}

	// A dedicated connection holds the session lock for the whole test.
	lockConn, err := sqlDB.Conn(context.Background())
	if err != nil {
		t.Fatalf("pgtest: lock connection: %v", err)
	}
	if _, err := lockConn.ExecContext(context.Background(), "SELECT pg_advisory_lock($1)", lockKey); err != nil {
		t.Fatalf("pgtest: acquire database lock: %v", err)
	}

	truncate(t, db)
	t.Cleanup(func() {
		truncate(t, db)
		_, _ = lockConn.ExecContext(context.Background(), "SELECT pg_advisory_unlock($1)", lockKey)
		_ = lockConn.Close()
		_ = sqlDB.Close()
	})

	return db
}

func truncate(t testing.TB, db *gorm.DB) {
	t.Helper()

	if err := db.Exec("TRUNCATE TABLE asset, asset_import").Error; err != nil {
		t.Fatalf("pgtest: truncate tables: %v", err)
	}
}
