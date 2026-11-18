package usecase

import (
	"context"
	"errors"
	"io"
	"strings"

	"smart-grid-asset-registry/domain/assetimport"
	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/valueobject"
)

// AssetImportUsecase: one method = one endpoint = one file in
// internal/usecase/asset_import/. There are no users or permissions
// (constitution, Template Exception 1), so inputs carry no caller identity.
type AssetImportUsecase interface {
	// ValidateAssetImport checks a file against the store; nothing is stored.
	ValidateAssetImport(ctx context.Context, input ValidateAssetImportInput) (ValidateAssetImportOutput, error)
	// CreateAssetImport re-validates a file and applies it atomically.
	CreateAssetImport(ctx context.Context, input CreateAssetImportInput) (CreateAssetImportOutput, error)
}

const (
	// SystemActorName is the only value written to *_by_name audit columns
	// while the product has no users.
	SystemActorName = "system"
	// MaxFileBytes is the largest accepted CSV file (5 MB).
	MaxFileBytes = 5 * 1024 * 1024
)

// ---- ValidateAssetImport ----------------------------------------------------

type ValidateAssetImportInput struct {
	FileName string
	File     io.Reader
	Mode     valueobject.ImportMode
}

type ValidateAssetImportOutput struct {
	Report assetimport.Report
}

// Validate checks the request shape only. The file content is checked by the
// import rules; the size limit is enforced while parsing.
func (i ValidateAssetImportInput) Validate() error {
	return validateImportRequest(i.FileName, i.File, i.Mode)
}

// ---- CreateAssetImport ------------------------------------------------------

type CreateAssetImportInput struct {
	FileName string
	File     io.Reader
	Mode     valueobject.ImportMode
}

type CreateAssetImportOutput struct {
	Import *entity.AssetImport
	Report assetimport.Report
}

func (i CreateAssetImportInput) Validate() error {
	return validateImportRequest(i.FileName, i.File, i.Mode)
}

// ---- shared -----------------------------------------------------------------

func validateImportRequest(fileName string, file io.Reader, mode valueobject.ImportMode) error {
	if file == nil {
		return errors.New("file is required")
	}
	if strings.TrimSpace(fileName) == "" {
		return errors.New("file name is required")
	}
	return mode.IsValid()
}
