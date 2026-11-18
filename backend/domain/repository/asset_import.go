package repository

import (
	"context"

	"smart-grid-asset-registry/domain/entity"
)

// AssetImportRepository stores import records and serialises imports.
type AssetImportRepository interface {
	Create(ctx context.Context, trx Tx, assetImport *entity.AssetImport) error
	// AcquireLock takes the transaction-scoped advisory lock that lets only
	// one import run at a time (spec FR-024). trx is required: the lock is
	// released when that transaction commits or rolls back.
	AcquireLock(ctx context.Context, trx Tx) error
}
