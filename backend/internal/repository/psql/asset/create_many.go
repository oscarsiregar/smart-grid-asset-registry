package asset_repository

import (
	"context"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/repository"
	"smart-grid-asset-registry/internal/repository/psql/gormtx"
	"smart-grid-asset-registry/internal/repository/psql/mapper"
)

// createBatchSize keeps one INSERT at 1 000 rows × 17 columns, far below
// PostgreSQL's 65 535 bind-parameter limit (research R6).
const createBatchSize = 1000

// CreateMany inserts the assets in the given order, batch by batch. Callers
// pass parents before children so the parent foreign key holds.
func (r assetRepository) CreateMany(ctx context.Context, trx repository.Tx, assets []*entity.Asset) error {
	if len(assets) == 0 {
		return nil
	}

	db, err := gormtx.Conn(ctx, r.db, trx)
	if err != nil {
		return err
	}

	return db.CreateInBatches(mapper.ToModelAssets(assets), createBatchSize).Error
}
