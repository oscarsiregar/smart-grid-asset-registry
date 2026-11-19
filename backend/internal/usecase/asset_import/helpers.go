package asset_import_usecase

import (
	"context"
	"io"

	"smart-grid-asset-registry/domain/assetimport"
	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/errorcode"
	"smart-grid-asset-registry/domain/repository"
	"smart-grid-asset-registry/domain/usecase"
	"smart-grid-asset-registry/internal/apperr"
)

// validatable is the request-shape check every import input has.
type validatable interface {
	Validate() error
}

// parseInput checks the request shape (400 INVALID_REQUEST) and parses the
// file; a file the rules refuse as a whole is 422 with its file-level code
// (spec FR-003). assetimport.Parse is the only place the 5 MB limit is
// enforced.
func parseInput(input validatable, file io.Reader) (assetimport.ParsedFile, error) {
	if err := input.Validate(); err != nil {
		return assetimport.ParsedFile{}, apperr.BadRequest(errorcode.InvalidRequest, err.Error())
	}

	parsed, fileErr := assetimport.Parse(file, usecase.MaxFileBytes)
	if fileErr != nil {
		return assetimport.ParsedFile{}, apperr.UnprocessableEntity(fileErr.Code, fileErr.Args...)
	}
	return parsed, nil
}

// loadStoredAssets reads the stored assets the file mentions — by asset or
// parent ID, and by serial number — with one query each (research R6).
func (u assetImportUsecase) loadStoredAssets(ctx context.Context, trx repository.Tx, file assetimport.ParsedFile) (assetimport.StoredAssets, error) {
	byID, err := u.assetRepository.FindByAssetIDs(ctx, trx, file.ReferencedAssetIDs())
	if err != nil {
		return assetimport.StoredAssets{}, err
	}
	bySerial, err := u.assetRepository.FindBySerialNumbers(ctx, trx, file.SerialNumbers())
	if err != nil {
		return assetimport.StoredAssets{}, err
	}

	stored := assetimport.StoredAssets{
		ByAssetID: make(map[string]*entity.Asset, len(byID)),
		BySerial:  make(map[string][]string, len(bySerial)),
	}
	for _, asset := range byID {
		stored.ByAssetID[asset.AssetID] = asset
	}
	for _, asset := range bySerial {
		if asset.SerialNumber != nil {
			stored.BySerial[*asset.SerialNumber] = append(stored.BySerial[*asset.SerialNumber], asset.AssetID)
		}
	}
	return stored, nil
}
