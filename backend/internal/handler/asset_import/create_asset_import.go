package asset_import

import (
	"smart-grid-asset-registry/domain/usecase"
	"smart-grid-asset-registry/pkg/gocommon/response"

	"github.com/gofiber/fiber/v2"
)

// CreateAssetImport godoc
// POST /api/imports (multipart/form-data: file, mode)
//
// Every completed attempt stores an import record (spec FR-029), so the
// answer is 201 for COMMITTED, NOTHING_TO_COMMIT and STRICT_BLOCKED alike;
// isCommitted says whether assets were saved.
func (h *assetImportHandler) CreateAssetImport(ctx *fiber.Ctx) error {
	request, err := readUpload(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = request.file.Close() }()

	output, err := h.assetImportUsecase.CreateAssetImport(ctx.UserContext(), usecase.CreateAssetImportInput{
		FileName: request.fileName,
		File:     request.file,
		Mode:     request.mode,
	})
	if err != nil {
		return err
	}

	return response.Success(ctx, fiber.StatusCreated, "Successfully create asset import",
		toImportResultResponse(output.Import.Outcome, request.mode, request.fileName, output.Report, output.Import))
}
