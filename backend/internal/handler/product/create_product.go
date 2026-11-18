package product

import (
	"backend-template/domain/dto"
	"backend-template/domain/errorcode"
	"backend-template/domain/usecase"
	"backend-template/domain/valueobject"
	"backend-template/internal/apperr"

	"github.com/gofiber/fiber/v2"
	"repo.smf-indonesia.co.id/package-helper/go-common/response"
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
