package asset_import

import (
	"smart-grid-asset-registry/domain/usecase"
	"smart-grid-asset-registry/domain/valueobject"
	"smart-grid-asset-registry/pkg/gocommon/response"

	"github.com/gofiber/fiber/v2"
)

// ValidateAssetImport godoc
// POST /api/imports/validate (multipart/form-data: file, mode)
func (h *assetImportHandler) ValidateAssetImport(ctx *fiber.Ctx) error {
	request, err := readUpload(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = request.file.Close() }()

	output, err := h.assetImportUsecase.ValidateAssetImport(ctx.UserContext(), usecase.ValidateAssetImportInput{
		FileName: request.fileName,
		File:     request.file,
		Mode:     request.mode,
	})
	if err != nil {
		return err
	}

	return response.Success(ctx, fiber.StatusOK, "Successfully validate asset import",
		toImportResultResponse(valueobject.ImportOutcomeValidatedOnly, request.mode, request.fileName, output.Report, nil))
}
