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

// GetListProduct godoc
// GET /v1/products?search=&status=ACTIVE|INACTIVE&page=1&limit=10&sort=created_at&order=DESC
func (h *productHandler) GetListProduct(ctx *fiber.Ctx) error {
	var request dto.GetListProductRequest
	if err := ctx.QueryParser(&request); err != nil {
		return apperr.BadRequest(errorcode.InvalidRequest, err.Error())
	}

	var status *valueobject.ProductStatus
	if request.Status != "" {
		parsed, err := valueobject.ParseProductStatus(request.Status)
		if err != nil {
			return apperr.BadRequest(errorcode.InvalidRequest, err.Error())
		}
		status = &parsed
	}

	_, permissions := caller(ctx)

	output, err := h.productUsecase.GetListProduct(ctx.UserContext(), usecase.GetListProductInput{
		Search:      request.Search,
		Status:      status,
		Page:        request.Page,
		Limit:       request.Limit,
		Sort:        request.Sort,
		Order:       request.Order,
		Permissions: permissions,
	})
	if err != nil {
		return err
	}

	items := make([]dto.ProductResponse, 0, len(output.Items))
	for _, p := range output.Items {
		items = append(items, toProductResponse(p))
	}

	return response.Success(ctx, fiber.StatusOK, "Successfully get list product", dto.GetListProductResponse{
		Items:      items,
		TotalItems: output.TotalItems,
	})
}
