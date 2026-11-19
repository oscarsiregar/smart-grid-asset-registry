package asset_repository

import (
	"context"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/repository"
	"smart-grid-asset-registry/internal/repository/psql/gormtx"
	"smart-grid-asset-registry/internal/repository/psql/mapper"
	"smart-grid-asset-registry/internal/repository/psql/model"
)

func (r assetRepository) FindByAssetIDs(ctx context.Context, trx repository.Tx, assetIDs []string) ([]*entity.Asset, error) {
	if len(assetIDs) == 0 {
		return []*entity.Asset{}, nil
	}

	db, err := gormtx.Conn(ctx, r.db, trx)
	if err != nil {
		return nil, err
	}

	var rows []*model.Asset
	if err := db.Where("asset_id = ANY(?)", textArray(assetIDs)).Order("asset_id").Find(&rows).Error; err != nil {
		return nil, err
	}
	return mapper.ToEntityAssets(rows), nil
}
