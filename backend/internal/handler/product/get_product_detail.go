package product

import (
	"backend-template/domain/dto"
	"backend-template/domain/errorcode"
	"backend-template/domain/usecase"
	"backend-template/internal/apperr"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"repo.smf-indonesia.co.id/package-helper/go-common/response"
)

// GetProductDetail godoc
// GET /v1/products/:id
func (h *productHandler) GetProductDetail(ctx *fiber.Ctx) error {
	var request dto.GetProductDetailRequest
	if err := ctx.ParamsParser(&request); err != nil {
		return apperr.BadRequest(errorcode.InvalidRequest, err.Error())
	}

	id, err := uuid.Parse(request.ID)
	if err != nil {
		return apperr.BadRequest(errorcode.InvalidRequest, "id must be a valid UUID")
	}

	_, permissions := caller(ctx)

	output, err := h.productUsecase.GetProductDetail(ctx.UserContext(), usecase.GetProductDetailInput{
		ID:          id,
		Permissions: permissions,
	})
	if err != nil {
		return err
	}

	return response.Success(ctx, fiber.StatusOK, "Successfully get product detail", toProductResponse(output.Product))
}
