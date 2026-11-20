package routes

import (
	"smart-grid-asset-registry/internal/config"
	asset "smart-grid-asset-registry/internal/handler/asset"
	asset_import "smart-grid-asset-registry/internal/handler/asset_import"
	"smart-grid-asset-registry/pkg/gocommon/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type RouteDependencies struct {
	App                *fiber.App
	DB                 *gorm.DB
	Cfg                *config.Config
	AssetImportHandler asset_import.AssetImportHandler
	AssetHandler       asset.AssetHandler
}

// RegisterRoutes is the only place routes are registered.
//
// Every route is public and lives under /api (constitution, Template
// Exceptions 1 and 5): the app has no users, so the template's
// InternalSecret → Identity → JWTVerifyStrict chain does not apply. If users
// are ever introduced, restore that chain (backend/docs/rules/07-security.md).
func RegisterRoutes(deps *RouteDependencies) {
	api := deps.App.Group("/api")

	// Liveness/readiness probe (pings the DB).
	utils.RegisterHealthCheck(api, deps.DB, nil)

	// Static routes are registered before any parameterised route.
	importGroup := api.Group("/imports")
	importGroup.Post("/validate", deps.AssetImportHandler.ValidateAssetImport)
	importGroup.Post("/", deps.AssetImportHandler.CreateAssetImport)

	// Read-only asset explorer (feature 002); complete lists, no paging
	// (Template Exception 13).
	assetGroup := api.Group("/assets")
	assetGroup.Get("/", deps.AssetHandler.GetListAsset)
	assetGroup.Get("/:assetId", deps.AssetHandler.GetAssetDetail)
}
