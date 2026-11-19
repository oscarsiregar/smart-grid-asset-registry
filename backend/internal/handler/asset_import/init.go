package asset_import

import (
	"smart-grid-asset-registry/domain/usecase"

	"github.com/gofiber/fiber/v2"
)

// AssetImportHandler lists every endpoint of the module; routes.go wires them.
type AssetImportHandler interface {
	ValidateAssetImport(ctx *fiber.Ctx) error
}

type assetImportHandler struct {
	assetImportUsecase usecase.AssetImportUsecase
}

func NewAssetImportHandler(assetImportUsecase usecase.AssetImportUsecase) AssetImportHandler {
	return &assetImportHandler{
		assetImportUsecase: assetImportUsecase,
	}
}
