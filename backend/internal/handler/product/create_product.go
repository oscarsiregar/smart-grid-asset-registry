package product

import (
	"smart-grid-asset-registry/domain/dto"
	"smart-grid-asset-registry/domain/errorcode"
	"smart-grid-asset-registry/domain/usecase"
	"smart-grid-asset-registry/domain/valueobject"
	"smart-grid-asset-registry/internal/apperr"
	"smart-grid-asset-registry/pkg/gocommon/response"

	"github.com/gofiber/fiber/v2"
)

// CreateProduct godoc
// POST /v1/products
func (h *productHandler) CreateProduct(ctx *fiber.Ctx) error {
	var request dto.CreateProductRequest
	if err := ctx.BodyParser(&request); err != nil {
		return apperr.BadRequest(errorcode.InvalidRequest, err.Error())
	}

	status, err := valueobject.ParseProductStatus(request.Status)
	if err != nil {
		return apperr.BadRequest(errorcode.InvalidRequest, err.Error())
	}

	actorName, permissions := caller(ctx)

	output, err := h.productUsecase.CreateProduct(ctx.UserContext(), usecase.CreateProductInput{
		Code:        request.Code,
		Name:        request.Name,
		Description: request.Description,
		Price:       request.Price,
		Status:      status,
		ActorName:   actorName,
		Permissions: permissions,
	})
	if err != nil {
		return err
	}

	return response.Success(ctx, fiber.StatusCreated, "Successfully create product", toProductResponse(output.Product))
}
