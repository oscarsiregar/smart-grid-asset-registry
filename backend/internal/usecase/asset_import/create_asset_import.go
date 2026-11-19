package asset_import_usecase

import (
	"context"
	"errors"

	"smart-grid-asset-registry/domain/usecase"
)

// errCreateNotImplemented marks the placeholder below.
var errCreateNotImplemented = errors.New("create asset import: not implemented yet")

// CreateAssetImport is a PLACEHOLDER until task T082 (user story 2) replaces
// it with the real commit flow. Its route (POST /api/imports) is registered
// only by T084, so nothing can reach this method before then.
func (u assetImportUsecase) CreateAssetImport(context.Context, usecase.CreateAssetImportInput) (usecase.CreateAssetImportOutput, error) {
	return usecase.CreateAssetImportOutput{}, errCreateNotImplemented
}
