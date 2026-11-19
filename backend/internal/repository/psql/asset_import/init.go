package asset_import_repository

import (
	"smart-grid-asset-registry/domain/repository"

	"gorm.io/gorm"
)

// importLockKey identifies the transaction-scoped advisory lock that lets
// only one import run at a time (spec FR-024, research R5). Any fixed number
// works as long as nothing else in the database uses it.
const importLockKey int64 = 7_334_455_002

type assetImportRepository struct {
	db *gorm.DB
}

func NewAssetImportRepository(db *gorm.DB) repository.AssetImportRepository {
	return &assetImportRepository{
		db: db,
	}
}
