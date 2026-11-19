package asset_import_repository

import (
	"testing"

	"smart-grid-asset-registry/internal/repository/psql/gormtx"

	"github.com/DATA-DOG/go-sqlmock"
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

// wrapTx wraps a mock *gorm.DB into the concrete *gormtx.Tx the real code
// unwraps, so the trx-provided branch runs against the same expectations.
func wrapTx(db *gorm.DB) *gormtx.Tx {
	return gormtx.New(db)
}

// unsupportedTx satisfies repository.Tx but is not a *gormtx.Tx.
type unsupportedTx struct{}

func (unsupportedTx) Commit() error   { return nil }
func (unsupportedTx) Rollback() error { return nil }
