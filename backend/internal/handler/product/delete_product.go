package product

import (
	"backend-template/domain/dto"
	"backend-template/domain/errorcode"
	"backend-template/domain/usecase"
	"backend-template/internal/apperr"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// DeleteProduct godoc
// DELETE /v1/products/:id  -> 204 No Content
func (h *productHandler) DeleteProduct(ctx *fiber.Ctx) error {
	var request dto.DeleteProductRequest
	if err := ctx.ParamsParser(&request); err != nil {
		return apperr.BadRequest(errorcode.InvalidRequest, err.Error())
	}

	id, err := uuid.Parse(request.ID)
	if err != nil {
		return apperr.BadRequest(errorcode.InvalidRequest, "id must be a valid UUID")
	}

	_, permissions := caller(ctx)

	if err := h.productUsecase.DeleteProduct(ctx.UserContext(), usecase.DeleteProductInput{
		ID:          id,
		Permissions: permissions,
	}); err != nil {
		return err
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}
