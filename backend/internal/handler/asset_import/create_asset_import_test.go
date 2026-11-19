package asset_import

import (
	"errors"
	"io"
	"testing"
	"time"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/errorcode"
	"smart-grid-asset-registry/domain/usecase"
	"smart-grid-asset-registry/domain/valueobject"
	"smart-grid-asset-registry/internal/apperr"
	mocksusecase "smart-grid-asset-registry/internal/mocks/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const createTarget = "/api/imports"

func storedRecord(outcome valueobject.ImportOutcome) *entity.AssetImport {
	return &entity.AssetImport{
		ID:        uuid.MustParse("6f1c2c1e-6a4e-4b8e-9a52-3d1f2b9c0a11"),
		Outcome:   outcome,
		CreatedAt: time.Date(2026, 11, 19, 10, 15, 0, 0, time.UTC),
	}
}

func TestAssetImportHandler_CreateAssetImport(t *testing.T) {
	csv := "asset_id\nSS-1\n"

	t.Run("forwards the file and answers 201 with the record and isCommitted", func(t *testing.T) {
		uc := mocksusecase.NewAssetImportUsecase(t)
		uc.EXPECT().CreateAssetImport(mock.Anything, mock.MatchedBy(func(in usecase.CreateAssetImportInput) bool {
			content, err := io.ReadAll(in.File)
			return err == nil && string(content) == csv && in.FileName == "grid_assets.csv" &&
				in.Mode == valueobject.ImportModePartial
		})).Return(usecase.CreateAssetImportOutput{
			Import: storedRecord(valueobject.ImportOutcomeCommitted),
			Report: sampleReport(),
		}, nil)

		resp := doMultipart(t, newTestApp(NewAssetImportHandler(uc)), createTarget, filePart(csv))

		require.Equal(t, fiber.StatusCreated, resp.StatusCode)
		body := decodeEnvelope(t, resp.Body)
		assert.Equal(t, "Successfully create asset import", body.Message)
		assert.Contains(t, string(body.Data), `"outcome":"COMMITTED"`)
		assert.Contains(t, string(body.Data), `"isCommitted":true`)
		assert.Contains(t, string(body.Data), `"import":{"id":"6f1c2c1e-6a4e-4b8e-9a52-3d1f2b9c0a11","createdAt":"2026-11-19T10:15:00Z"}`)
	})

	t.Run("a completed attempt that saved nothing is still 201, not committed", func(t *testing.T) {
		uc := mocksusecase.NewAssetImportUsecase(t)
		uc.EXPECT().CreateAssetImport(mock.Anything, mock.Anything).Return(usecase.CreateAssetImportOutput{
			Import: storedRecord(valueobject.ImportOutcomeNothingToCommit),
		}, nil)

		resp := doMultipart(t, newTestApp(NewAssetImportHandler(uc)), createTarget, filePart(csv), modePart("STRICT"))

		require.Equal(t, fiber.StatusCreated, resp.StatusCode)
		body := decodeEnvelope(t, resp.Body)
		assert.Contains(t, string(body.Data), `"outcome":"NOTHING_TO_COMMIT"`)
		assert.Contains(t, string(body.Data), `"isCommitted":false`)
		assert.Contains(t, string(body.Data), `"mode":"STRICT"`)
	})

	t.Run("missing file part is 400 and the usecase is not called", func(t *testing.T) {
		uc := mocksusecase.NewAssetImportUsecase(t)

		resp := doMultipart(t, newTestApp(NewAssetImportHandler(uc)), createTarget, modePart("PARTIAL"))

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("a usecase 422 passes through with its error code", func(t *testing.T) {
		uc := mocksusecase.NewAssetImportUsecase(t)
		uc.EXPECT().CreateAssetImport(mock.Anything, mock.Anything).
			Return(usecase.CreateAssetImportOutput{}, apperr.UnprocessableEntity(errorcode.ImportFileEmpty))

		resp := doMultipart(t, newTestApp(NewAssetImportHandler(uc)), createTarget, filePart(csv))

		require.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)
		assert.JSONEq(t, `{"errorCode":"IMPORT_FILE_EMPTY"}`, string(decodeEnvelope(t, resp.Body).Data))
	})

	t.Run("a technical error (rolled back) is a 500 without details", func(t *testing.T) {
		uc := mocksusecase.NewAssetImportUsecase(t)
		uc.EXPECT().CreateAssetImport(mock.Anything, mock.Anything).
			Return(usecase.CreateAssetImportOutput{}, errors.New("pq: connection refused"))

		resp := doMultipart(t, newTestApp(NewAssetImportHandler(uc)), createTarget, filePart(csv))

		require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		assert.Equal(t, "Internal Server Error", decodeEnvelope(t, resp.Body).Message)
	})
}
