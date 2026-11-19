package asset_import_repository

import (
	"context"
	"errors"

	"smart-grid-asset-registry/domain/repository"
	"smart-grid-asset-registry/internal/repository/psql/gormtx"
)

var errLockNeedsTransaction = errors.New("asset import lock: a transaction is required")

// AcquireLock waits for the import advisory lock. The lock is transaction
// scoped (pg_advisory_xact_lock), so it is released by the commit or rollback
// of trx; without a transaction it would be released at once.
func (r assetImportRepository) AcquireLock(ctx context.Context, trx repository.Tx) error {
	if trx == nil {
		return errLockNeedsTransaction
	}

	db, err := gormtx.Conn(ctx, r.db, trx)
	if err != nil {
		return err
	}

	return db.Exec("SELECT pg_advisory_xact_lock(?)", importLockKey).Error
}
