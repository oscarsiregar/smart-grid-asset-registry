package asset_import_usecase

import (
	"context"

	"smart-grid-asset-registry/domain/assetimport"
	"smart-grid-asset-registry/domain/repository"
	"smart-grid-asset-registry/domain/usecase"
	"smart-grid-asset-registry/pkg/gocommon/pkg/logger"

	"github.com/sirupsen/logrus"
)

// ValidateAssetImport checks a file against the current store and stores
// nothing: no transaction, no lock, no import record (spec FR-001).
func (u assetImportUsecase) ValidateAssetImport(ctx context.Context, input usecase.ValidateAssetImportInput) (usecase.ValidateAssetImportOutput, error) {
	log := logger.WithCtx(ctx, u.log).WithField("usecase", "ValidateAssetImport")

	file, err := parseInput(input, input.File)
	if err != nil {
		return usecase.ValidateAssetImportOutput{}, err
	}

	var noTransaction repository.Tx
	stored, err := u.loadStoredAssets(ctx, noTransaction, file)
	if err != nil {
		log.WithError(err).Error("failed to load stored assets")
		return usecase.ValidateAssetImportOutput{}, err
	}

	report := assetimport.Validate(file, stored)

	log.WithFields(reportLogFields(input.FileName, input.Mode.String(), report)).Info("success")
	return usecase.ValidateAssetImportOutput{Report: report}, nil
}

// reportLogFields are the fields every import usecase logs on success:
// counts and identifiers only, never row contents.
func reportLogFields(fileName, mode string, report assetimport.Report) logrus.Fields {
	return logrus.Fields{
		"fileName":    fileName,
		"mode":        mode,
		"total":       report.Counts.Total,
		"toImport":    report.Counts.ToImport,
		"unchanged":   report.Counts.Unchanged,
		"rejected":    report.Counts.Rejected,
		"warningRows": report.Counts.WarningRows,
	}
}
