package asset_usecase

import (
	"smart-grid-asset-registry/domain/repository"
	"smart-grid-asset-registry/domain/usecase"

	"github.com/sirupsen/logrus"
)

// assetUsecase only reads, so it needs no transaction manager (research R5).
type assetUsecase struct {
	assetRepository repository.AssetRepository
	log             *logrus.Logger
}

func NewAssetUsecase(assetRepository repository.AssetRepository, log *logrus.Logger) usecase.AssetUsecase {
	return &assetUsecase{
		assetRepository: assetRepository,
		log:             log,
	}
}
