package asset_import

import (
	"errors"
	"io"
	"testing"

	"smart-grid-asset-registry/domain/errorcode"
	"smart-grid-asset-registry/domain/usecase"
	"smart-grid-asset-registry/domain/valueobject"
	"smart-grid-asset-registry/internal/apperr"
	mocksusecase "smart-grid-asset-registry/internal/mocks/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const validateTarget = "/api/imports/validate"

func TestAssetImportHandler_ValidateAssetImport(t *testing.T) {
	csv := "asset_id\nSS-1\n"

	t.Run("forwards the file, name and mode and answers 200 VALIDATED_ONLY", func(t *testing.T) {
		uc := mocksusecase.NewAssetImportUsecase(t)
		uc.EXPECT().ValidateAssetImport(mock.Anything, mock.MatchedBy(func(in usecase.ValidateAssetImportInput) bool {
			content, err := io.ReadAll(in.File)
			return err == nil && string(content) == csv &&
				in.FileName == "grid_assets.csv" && in.Mode == valueobject.ImportModeStrict
		})).Return(usecase.ValidateAssetImportOutput{Report: sampleReport()}, nil)

		resp := doMultipart(t, newTestApp(NewAssetImportHandler(uc)), validateTarget, filePart(csv), modePart("STRICT"))

		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		body := decodeEnvelope(t, resp.Body)
		assert.Equal(t, "Successfully validate asset import", body.Message)
		assert.Contains(t, string(body.Data), `"outcome":"VALIDATED_ONLY"`)
		assert.Contains(t, string(body.Data), `"isCommitted":false`)
		assert.Contains(t, string(body.Data), `"mode":"STRICT"`)
		assert.Contains(t, string(body.Data), `"import":null`)
	})

	t.Run("mode defaults to PARTIAL", func(t *testing.T) {
		uc := mocksusecase.NewAssetImportUsecase(t)
		uc.EXPECT().ValidateAssetImport(mock.Anything, mock.MatchedBy(func(in usecase.ValidateAssetImportInput) bool {
			return in.Mode == valueobject.ImportModePartial
		})).Return(usecase.ValidateAssetImportOutput{}, nil)

		resp := doMultipart(t, newTestApp(NewAssetImportHandler(uc)), validateTarget, filePart(csv))

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("missing file part is 400 INVALID_REQUEST and the usecase is not called", func(t *testing.T) {
		uc := mocksusecase.NewAssetImportUsecase(t)

		resp := doMultipart(t, newTestApp(NewAssetImportHandler(uc)), validateTarget, modePart("PARTIAL"))

		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		assert.JSONEq(t, `{"errorCode":"INVALID_REQUEST"}`, string(decodeEnvelope(t, resp.Body).Data))
	})

	t.Run("unknown mode is 400 INVALID_REQUEST and the usecase is not called", func(t *testing.T) {
		uc := mocksusecase.NewAssetImportUsecase(t)

		resp := doMultipart(t, newTestApp(NewAssetImportHandler(uc)), validateTarget, filePart(csv), modePart("ALL"))

		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		body := decodeEnvelope(t, resp.Body)
		assert.Equal(t, "Invalid request: mode must be PARTIAL or STRICT", body.Message)
	})

	t.Run("a usecase 422 passes through with its error code", func(t *testing.T) {
		uc := mocksusecase.NewAssetImportUsecase(t)
		uc.EXPECT().ValidateAssetImport(mock.Anything, mock.Anything).
			Return(usecase.ValidateAssetImportOutput{}, apperr.UnprocessableEntity(errorcode.ImportMissingColumns, "asset_name"))

		resp := doMultipart(t, newTestApp(NewAssetImportHandler(uc)), validateTarget, filePart(csv))

		require.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)
		body := decodeEnvelope(t, resp.Body)
		assert.Equal(t, "The file is missing required columns: asset_name.", body.Message)
		assert.JSONEq(t, `{"errorCode":"IMPORT_MISSING_COLUMNS"}`, string(body.Data))
	})

	t.Run("a technical error is a 500 without details", func(t *testing.T) {
		uc := mocksusecase.NewAssetImportUsecase(t)
		uc.EXPECT().ValidateAssetImport(mock.Anything, mock.Anything).
			Return(usecase.ValidateAssetImportOutput{}, errors.New("pq: connection refused"))

		resp := doMultipart(t, newTestApp(NewAssetImportHandler(uc)), validateTarget, filePart(csv))

		require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		body := decodeEnvelope(t, resp.Body)
		assert.Equal(t, "Internal Server Error", body.Message)
		assert.Equal(t, "null", string(body.Data))
	})
}
