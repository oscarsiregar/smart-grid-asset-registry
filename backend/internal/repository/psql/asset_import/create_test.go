package asset_import_repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/valueobject"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const insertSQL = `INSERT INTO "asset_import" \("id","file_name","mode","outcome","total_rows","imported_rows",` +
	`"unchanged_rows","rejected_rows","warning_rows","created_at","created_by_name","updated_at","updated_by_name"\) ` +
	`VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7,\$8,\$9,\$10,\$11,\$12,\$13\)$`

func newRecord() *entity.AssetImport {
	now := time.Date(2026, time.November, 19, 10, 0, 0, 0, time.UTC)
	return &entity.AssetImport{
		ID: uuid.New(), FileName: "grid.csv", Mode: valueobject.ImportModePartial, Outcome: valueobject.ImportOutcomeCommitted,
		TotalRows: 222, ImportedRows: 205, UnchangedRows: 0, RejectedRows: 17, WarningRows: 2,
		CreatedAt: now, CreatedByName: "system", UpdatedAt: now, UpdatedByName: "system",
	}
}

func TestAssetImportRepository_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("inserts every column", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		record := newRecord()
		mock.ExpectBegin()
		mock.ExpectExec(insertSQL).
			WithArgs(record.ID, "grid.csv", valueobject.ImportModePartial, valueobject.ImportOutcomeCommitted,
				222, 205, 0, 17, 2, record.CreatedAt, "system", record.UpdatedAt, "system").
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		assert.NoError(t, NewAssetImportRepository(gdb).Create(ctx, nil, record))
	})

	t.Run("propagates a database error", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		dbErr := errors.New("connection refused")
		mock.ExpectBegin()
		mock.ExpectExec(insertSQL).WillReturnError(dbErr)
		mock.ExpectRollback()

		assert.ErrorIs(t, NewAssetImportRepository(gdb).Create(ctx, nil, newRecord()), dbErr)
	})

	t.Run("uses the given transaction", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		mock.ExpectBegin()
		mock.ExpectExec(insertSQL).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		assert.NoError(t, NewAssetImportRepository(gdb).Create(ctx, wrapTx(gdb), newRecord()))
	})

	t.Run("returns an error for an unsupported Tx type", func(t *testing.T) {
		gdb, _ := newMockDB(t)

		assert.Error(t, NewAssetImportRepository(gdb).Create(ctx, unsupportedTx{}, newRecord()))
	})
}
