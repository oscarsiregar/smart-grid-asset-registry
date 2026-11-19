package integrationtest

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/repository"
	"smart-grid-asset-registry/domain/valueobject"
	asset_repository "smart-grid-asset-registry/internal/repository/psql/asset"
	asset_import_repository "smart-grid-asset-registry/internal/repository/psql/asset_import"
	"smart-grid-asset-registry/internal/repository/psql/gormtx"
	"smart-grid-asset-registry/internal/testsupport/pgtest"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const lockWait = 300 * time.Millisecond

func newRecord(id uuid.UUID) *entity.AssetImport {
	return &entity.AssetImport{
		ID: id, FileName: "grid.csv", Mode: valueobject.ImportModePartial, Outcome: valueobject.ImportOutcomeCommitted,
		TotalRows: 2, ImportedRows: 2, CreatedAt: seedTime, CreatedByName: "system", UpdatedAt: seedTime, UpdatedByName: "system",
	}
}

func newAsset(importID uuid.UUID, assetID string, parentAssetID *string, assetType valueobject.AssetType) *entity.Asset {
	return &entity.Asset{
		ID: uuid.New(), AssetID: assetID, ParentAssetID: parentAssetID, AssetType: assetType, AssetName: assetID,
		ImportID: importID, CreatedAt: seedTime, CreatedByName: "system", UpdatedAt: seedTime, UpdatedByName: "system",
	}
}

func TestAssetImportRepository_WritesAgainstPostgreSQL(t *testing.T) {
	db := pgtest.Open(t)
	ctx := context.Background()
	txManager := gormtx.NewTransactionManager(db)
	imports := asset_import_repository.NewAssetImportRepository(db)
	assets := asset_repository.NewAssetRepository(db)

	importID := uuid.New()
	err := txManager.WithinTx(ctx, func(ctx context.Context, trx repository.Tx) error {
		if err := imports.AcquireLock(ctx, trx); err != nil {
			return err
		}
		if err := imports.Create(ctx, trx, newRecord(importID)); err != nil {
			return err
		}
		// Parents before children: the parent foreign key is checked per row.
		return assets.CreateMany(ctx, trx, []*entity.Asset{
			newAsset(importID, "SS-1", nil, valueobject.AssetTypeSubstation),
			newAsset(importID, "TX-1", pointer("SS-1"), valueobject.AssetTypeTransformer),
		})
	})

	require.NoError(t, err)
	assert.EqualValues(t, 1, countRows(t, db, "asset_import"))
	assert.EqualValues(t, 2, countRows(t, db, "asset"))
}

func TestAssetImportRepository_ForeignKeysAreEnforced(t *testing.T) {
	db := pgtest.Open(t)
	ctx := context.Background()
	assets := asset_repository.NewAssetRepository(db)

	t.Run("an asset needs its import record", func(t *testing.T) {
		err := assets.CreateMany(ctx, nil, []*entity.Asset{newAsset(uuid.New(), "SS-1", nil, valueobject.AssetTypeSubstation)})
		assert.Error(t, err)
	})

	t.Run("an asset needs its parent", func(t *testing.T) {
		importID := seedImport(t, db)
		err := assets.CreateMany(ctx, nil, []*entity.Asset{newAsset(importID, "TX-1", pointer("SS-404"), valueobject.AssetTypeTransformer)})
		assert.Error(t, err)
	})
}

// TestAssetImportRepository_LockSerialisesImports holds the lock in one
// transaction and checks that a second transaction waits for it.
func TestAssetImportRepository_LockSerialisesImports(t *testing.T) {
	db := pgtest.Open(t)
	ctx := context.Background()
	txManager := gormtx.NewTransactionManager(db)
	imports := asset_import_repository.NewAssetImportRepository(db)

	holding := make(chan struct{})
	release := make(chan struct{})
	firstDone := make(chan error, 1)
	go func() {
		firstDone <- txManager.WithinTx(ctx, func(ctx context.Context, trx repository.Tx) error {
			if err := imports.AcquireLock(ctx, trx); err != nil {
				return err
			}
			close(holding)
			<-release
			return nil
		})
	}()
	<-holding

	var secondAcquired atomic.Bool
	secondDone := make(chan error, 1)
	go func() {
		secondDone <- txManager.WithinTx(ctx, func(ctx context.Context, trx repository.Tx) error {
			if err := imports.AcquireLock(ctx, trx); err != nil {
				return err
			}
			secondAcquired.Store(true)
			return nil
		})
	}()

	time.Sleep(lockWait)
	assert.False(t, secondAcquired.Load(), "the second import must wait while the first holds the lock")

	close(release)
	require.NoError(t, <-firstDone)
	require.NoError(t, <-secondDone)
	assert.True(t, secondAcquired.Load(), "the second import runs once the first commits")
}
