package asset

import (
	"smart-grid-asset-registry/domain/usecase"

	"github.com/gofiber/fiber/v2"
)

// AssetHandler lists every endpoint of the module; routes.go wires them.
type AssetHandler interface {
	GetListAsset(ctx *fiber.Ctx) error
	GetAssetDetail(ctx *fiber.Ctx) error
}

type assetHandler struct {
	assetUsecase usecase.AssetUsecase
}

func NewAssetHandler(assetUsecase usecase.AssetUsecase) AssetHandler {
	return &assetHandler{
		assetUsecase: assetUsecase,
	}
}
