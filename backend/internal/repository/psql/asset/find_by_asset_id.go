package asset_repository

import (
	"context"
	"errors"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/repository"
	"smart-grid-asset-registry/internal/repository/psql/gormtx"
	"smart-grid-asset-registry/internal/repository/psql/mapper"
	"smart-grid-asset-registry/internal/repository/psql/model"

	"gorm.io/gorm"
)

func (r assetRepository) FindByAssetID(ctx context.Context, trx repository.Tx, assetID string) (*entity.Asset, error) {
	db, err := gormtx.Conn(ctx, r.db, trx)
	if err != nil {
		return nil, err
	}

	var row model.Asset
	if err := db.Where("asset_id = ?", assetID).Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.ToEntityAsset(&row), nil
}
