package asset

import (
	"smart-grid-asset-registry/domain/dto"
	"smart-grid-asset-registry/domain/errorcode"
	"smart-grid-asset-registry/domain/usecase"
	"smart-grid-asset-registry/internal/apperr"
	"smart-grid-asset-registry/pkg/gocommon/response"

	"github.com/gofiber/fiber/v2"
)

// GetAssetDetail: GET /api/assets/:assetId. Decoding a percent-encoded ID is
// added with user story 3 (research R2, tasks.md T065).
func (h assetHandler) GetAssetDetail(ctx *fiber.Ctx) error {
	var request dto.GetAssetDetailRequest
	if err := ctx.ParamsParser(&request); err != nil {
		return apperr.BadRequest(errorcode.InvalidRequest, "invalid path parameters")
	}

	output, err := h.assetUsecase.GetAssetDetail(ctx.UserContext(), usecase.GetAssetDetailInput{AssetID: request.AssetID})
	if err != nil {
		return err
	}

	return response.Success(ctx, fiber.StatusOK, "Successfully get asset detail", toAssetDetailResponse(output))
}
