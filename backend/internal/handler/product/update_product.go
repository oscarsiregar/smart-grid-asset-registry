package product

import (
	"smart-grid-asset-registry/domain/dto"
	"smart-grid-asset-registry/domain/errorcode"
	"smart-grid-asset-registry/domain/usecase"
	"smart-grid-asset-registry/domain/valueobject"
	"smart-grid-asset-registry/internal/apperr"
	"smart-grid-asset-registry/pkg/gocommon/response"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// UpdateProduct godoc
// PUT /v1/products/:id  (full replacement of the mutable fields)
func (h *productHandler) UpdateProduct(ctx *fiber.Ctx) error {
	var request dto.UpdateProductRequest
	if err := ctx.ParamsParser(&request); err != nil {
		return apperr.BadRequest(errorcode.InvalidRequest, err.Error())
	}
	if err := ctx.BodyParser(&request); err != nil {
		return apperr.BadRequest(errorcode.InvalidRequest, err.Error())
	}

	id, err := uuid.Parse(request.ID)
	if err != nil {
		return apperr.BadRequest(errorcode.InvalidRequest, "id must be a valid UUID")
	}

	status, err := valueobject.ParseProductStatus(request.Status)
	if err != nil {
		return apperr.BadRequest(errorcode.InvalidRequest, err.Error())
	}

	actorName, permissions := caller(ctx)

	output, err := h.productUsecase.UpdateProduct(ctx.UserContext(), usecase.UpdateProductInput{
		ID:          id,
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

	return response.Success(ctx, fiber.StatusOK, "Successfully update product", toProductResponse(output.Product))
}
