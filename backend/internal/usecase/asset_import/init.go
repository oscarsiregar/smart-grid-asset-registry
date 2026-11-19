package asset_import_usecase

import (
	"time"

	"smart-grid-asset-registry/domain/repository"
	"smart-grid-asset-registry/domain/usecase"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type assetImportUsecase struct {
	txManager             repository.TransactionManager
	assetRepository       repository.AssetRepository
	assetImportRepository repository.AssetImportRepository
	log                   *logrus.Logger

	// now and newID are injected so tests can pin timestamps and IDs.
	// Production: time.Now and uuid.New.
	now   func() time.Time
	newID func() uuid.UUID
}

func NewAssetImportUsecase(
	txManager repository.TransactionManager,
	assetRepository repository.AssetRepository,
	assetImportRepository repository.AssetImportRepository,
	log *logrus.Logger,
) usecase.AssetImportUsecase {
	return &assetImportUsecase{
		txManager:             txManager,
		assetRepository:       assetRepository,
		assetImportRepository: assetImportRepository,
		log:                   log,
		now:                   time.Now,
		newID:                 uuid.New,
	}
}
