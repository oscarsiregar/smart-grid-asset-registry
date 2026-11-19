package asset_repository

import (
	"smart-grid-asset-registry/domain/repository"

	"gorm.io/gorm"
)

type assetRepository struct {
	db *gorm.DB
}

func NewAssetRepository(db *gorm.DB) repository.AssetRepository {
	return &assetRepository{
		db: db,
	}
}
