package product

import (
	"smart-grid-asset-registry/domain/usecase"

	"github.com/gofiber/fiber/v2"
)

// ProductHandler lists every endpoint of the module; routes.go wires them.
type ProductHandler interface {
	GetListProduct(ctx *fiber.Ctx) error
	GetProductDetail(ctx *fiber.Ctx) error
	CreateProduct(ctx *fiber.Ctx) error
	UpdateProduct(ctx *fiber.Ctx) error
	DeleteProduct(ctx *fiber.Ctx) error
}

type productHandler struct {
	productUsecase usecase.ProductUsecase
}

func NewProductHandler(productUsecase usecase.ProductUsecase) ProductHandler {
	return &productHandler{
		productUsecase: productUsecase,
	}
}
