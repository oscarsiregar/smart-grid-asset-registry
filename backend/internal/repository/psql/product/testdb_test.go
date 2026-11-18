package product_repository

import (
	"database/sql/driver"
	"testing"
	"time"

	"smart-grid-asset-registry/internal/repository/psql/gormtx"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// newMockDB opens GORM on top of sqlmock (regexp matcher). Every test
// asserts mock.ExpectationsWereMet() via t.Cleanup.
func newMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, mock.ExpectationsWereMet())
		_ = sqlDB.Close()
	})

	gdb, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true,
	}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)

	return gdb, mock
}

func sqlmockNewResult() driver.Result {
	return sqlmock.NewResult(1, 1)
}

var productColumns = []string{"id", "code", "name", "description", "price", "status", "created_at", "created_by_name", "updated_at", "updated_by_name"}

func productRow(rows *sqlmock.Rows, id uuid.UUID, code string) *sqlmock.Rows {
	now := time.Now()
	return rows.AddRow(id, code, "Widget", nil, 10.5, "ACTIVE", now, "alice", now, "alice")
}

// wrapTx wraps a mock *gorm.DB into the concrete *gormtx.Tx the real code
// unwraps, so the trx-provided branch runs against the same expectations.
func wrapTx(db *gorm.DB) *gormtx.Tx {
	return gormtx.New(db)
}

// unsupportedTx satisfies repository.Tx but is not a *gormtx.Tx.
type unsupportedTx struct{}

func (unsupportedTx) Commit() error   { return nil }
func (unsupportedTx) Rollback() error { return nil }
