package asset_repository

import (
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

var assetColumns = []string{
	"id", "asset_id", "parent_asset_id", "asset_type", "asset_name", "operational_status", "voltage_kv",
	"rating_kva", "manufacturer", "model", "serial_number", "commissioned_date", "import_id",
	"created_at", "created_by_name", "updated_at", "updated_by_name",
}

func assetRow(rows *sqlmock.Rows, assetID string, parentAssetID any, assetType, serial string) *sqlmock.Rows {
	now := time.Now()
	return rows.AddRow(uuid.New(), assetID, parentAssetID, assetType, "Name "+assetID, "IN_SERVICE", 22.0,
		nil, nil, nil, serial, time.Date(2009, 2, 15, 0, 0, 0, 0, time.UTC), uuid.New(), now, "system", now, "system")
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

func pointer[T any](value T) *T { return &value }
