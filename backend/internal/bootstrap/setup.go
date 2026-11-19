// Package bootstrap is the composition root: the ONLY place concrete
// implementations are constructed and injected. Order: services ->
// repositories -> txManager -> usecases -> handlers.
package bootstrap

import (
	"smart-grid-asset-registry/internal/config"
	asset_import "smart-grid-asset-registry/internal/handler/asset_import"
	asset_repository "smart-grid-asset-registry/internal/repository/psql/asset"
	"smart-grid-asset-registry/internal/repository/psql/gormtx"
	"smart-grid-asset-registry/internal/routes"
	asset_import_usecase "smart-grid-asset-registry/internal/usecase/asset_import"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func SetupDependencies(app *fiber.App, db *gorm.DB, cfg *config.Config, log *logrus.Logger) *routes.RouteDependencies {
	// --- external services (internal/service/*) ---

	// --- repositories ---
	assetRepository := asset_repository.NewAssetRepository(db)

	txManager := gormtx.NewTransactionManager(db)

	// --- usecases ---
	// The asset-import repository (import records and the import lock) is
	// wired by task T084 together with POST /api/imports; until then only
	// ValidateAssetImport is reachable, and it never uses it.
	assetImportUsecase := asset_import_usecase.NewAssetImportUsecase(txManager, assetRepository, nil, log)

	// --- handlers ---
	assetImportHandler := asset_import.NewAssetImportHandler(assetImportUsecase)

	return &routes.RouteDependencies{
		App:                app,
		DB:                 db,
		Cfg:                cfg,
		AssetImportHandler: assetImportHandler,
	}
}
