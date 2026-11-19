package asset_import_repository

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

const lockSQL = `SELECT pg_advisory_xact_lock\(\$1\)$`

func TestAssetImportRepository_AcquireLock(t *testing.T) {
	ctx := context.Background()

	t.Run("takes the transaction-scoped advisory lock", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		mock.ExpectExec(lockSQL).WithArgs(importLockKey).WillReturnResult(sqlmock.NewResult(0, 0))

		assert.NoError(t, NewAssetImportRepository(gdb).AcquireLock(ctx, wrapTx(gdb)))
	})

	t.Run("propagates a database error", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		dbErr := errors.New("canceling statement due to lock timeout")
		mock.ExpectExec(lockSQL).WillReturnError(dbErr)

		assert.ErrorIs(t, NewAssetImportRepository(gdb).AcquireLock(ctx, wrapTx(gdb)), dbErr)
	})

	t.Run("refuses to run without a transaction", func(t *testing.T) {
		gdb, _ := newMockDB(t)

		assert.ErrorIs(t, NewAssetImportRepository(gdb).AcquireLock(ctx, nil), errLockNeedsTransaction)
	})

	t.Run("returns an error for an unsupported Tx type", func(t *testing.T) {
		gdb, _ := newMockDB(t)

		assert.Error(t, NewAssetImportRepository(gdb).AcquireLock(ctx, unsupportedTx{}))
	})
}
