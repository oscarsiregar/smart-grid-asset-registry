package asset_repository

import (
	"context"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/repository"
	"smart-grid-asset-registry/internal/repository/psql/gormtx"
	"smart-grid-asset-registry/internal/repository/psql/mapper"
	"smart-grid-asset-registry/internal/repository/psql/model"
)

func (r assetRepository) FindBySerialNumbers(ctx context.Context, trx repository.Tx, serialNumbers []string) ([]*entity.Asset, error) {
	if len(serialNumbers) == 0 {
		return []*entity.Asset{}, nil
	}

	db, err := gormtx.Conn(ctx, r.db, trx)
	if err != nil {
		return nil, err
	}

	var rows []*model.Asset
	if err := db.Where("serial_number = ANY(?)", textArray(serialNumbers)).Order("asset_id").Find(&rows).Error; err != nil {
		return nil, err
	}
	return mapper.ToEntityAssets(rows), nil
}
