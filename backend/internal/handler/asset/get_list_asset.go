package asset

import (
	"strings"

	"smart-grid-asset-registry/domain/dto"
	"smart-grid-asset-registry/domain/errorcode"
	"smart-grid-asset-registry/domain/usecase"
	"smart-grid-asset-registry/domain/valueobject"
	"smart-grid-asset-registry/internal/apperr"
	"smart-grid-asset-registry/pkg/gocommon/response"

	"github.com/gofiber/fiber/v2"
)

// GetListAsset: GET /api/assets[?parentAssetId=] or, with a non-blank q, the
// search (feature 003). Fiber decodes query values; q is trimmed, and a blank
// q means the hierarchy. Which parameters go together is the usecase's rule.
func (h assetHandler) GetListAsset(ctx *fiber.Ctx) error {
	var request dto.GetListAssetRequest
	if err := ctx.QueryParser(&request); err != nil {
		return apperr.BadRequest(errorcode.InvalidRequest, "invalid query parameters")
	}

	input := usecase.GetListAssetInput{
		Search: strings.TrimSpace(request.Q),
		Page:   request.Page,
		Limit:  request.Limit,
	}
	if request.ParentAssetID != "" {
		input.ParentAssetID = &request.ParentAssetID
	}
	if request.AssetType != "" {
		assetType, err := valueobject.ParseAssetType(request.AssetType)
		if err != nil {
			return apperr.BadRequest(errorcode.InvalidRequest, "assetType is not a valid asset type")
		}
		input.AssetType = &assetType
	}

	output, err := h.assetUsecase.GetListAsset(ctx.UserContext(), input)
	if err != nil {
		return err
	}

	if input.Search != "" {
		return response.Success(ctx, fiber.StatusOK, "Successfully get list asset", toAssetSearchResponse(output))
	}
	return response.Success(ctx, fiber.StatusOK, "Successfully get list asset", toAssetListResponse(output))
}
