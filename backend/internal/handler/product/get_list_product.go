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
