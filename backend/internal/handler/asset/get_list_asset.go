package asset

import (
	"smart-grid-asset-registry/domain/dto"
	"smart-grid-asset-registry/domain/errorcode"
	"smart-grid-asset-registry/domain/usecase"
	"smart-grid-asset-registry/internal/apperr"
	"smart-grid-asset-registry/pkg/gocommon/response"

	"github.com/gofiber/fiber/v2"
)

// GetListAsset: GET /api/assets[?parentAssetId=]. Fiber decodes the query
// value; an absent or empty parentAssetId means the top level.
func (h assetHandler) GetListAsset(ctx *fiber.Ctx) error {
	var request dto.GetListAssetRequest
	if err := ctx.QueryParser(&request); err != nil {
		return apperr.BadRequest(errorcode.InvalidRequest, "invalid query parameters")
	}

	input := usecase.GetListAssetInput{}
	if request.ParentAssetID != "" {
		input.ParentAssetID = &request.ParentAssetID
	}

	output, err := h.assetUsecase.GetListAsset(ctx.UserContext(), input)
	if err != nil {
		return err
	}

	return response.Success(ctx, fiber.StatusOK, "Successfully get list asset", toAssetListResponse(output))
}
