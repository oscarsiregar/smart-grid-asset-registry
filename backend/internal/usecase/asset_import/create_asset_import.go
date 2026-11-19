package asset_import_usecase

import (
	"context"

	"smart-grid-asset-registry/domain/assetimport"
	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/repository"
	"smart-grid-asset-registry/domain/usecase"
	"smart-grid-asset-registry/domain/valueobject"
	"smart-grid-asset-registry/pkg/gocommon/pkg/logger"

	"github.com/sirupsen/logrus"
)

// Steps of the import transaction, for the failure log.
const (
	stepBegin            = "begin transaction"
	stepAcquireLock      = "acquire import lock"
	stepLoadStoredAssets = "load stored assets"
	stepCreateRecord     = "create import record"
	stepCreateAssets     = "create assets"
	stepCommit           = "commit"
)

// CreateAssetImport re-validates the file against the store at this moment
// and applies it in ONE transaction (spec FR-022), serialised with every
// other import by an advisory lock (FR-024): lock → read the stored assets
// the file mentions → validate → store the import record → insert the
// accepted assets. Any technical error rolls the whole transaction back, so
// nothing is stored and no record is left; the failure is logged with the
// file name and mode (FR-030).
func (u assetImportUsecase) CreateAssetImport(ctx context.Context, input usecase.CreateAssetImportInput) (usecase.CreateAssetImportOutput, error) {
	log := logger.WithCtx(ctx, u.log).WithFields(logrus.Fields{
		"usecase":  "CreateAssetImport",
		"fileName": input.FileName,
		"mode":     input.Mode.String(),
	})

	file, err := parseInput(input, input.File)
	if err != nil {
		return usecase.CreateAssetImportOutput{}, err
	}

	var (
		record *entity.AssetImport
		report assetimport.Report
		// step names the part of the transaction that was running, so a
		// failure is logged once, with the file name and mode, wherever it
		// happened — including when the transaction cannot begin or commit.
		step = stepBegin
	)
	err = u.txManager.WithinTx(ctx, func(ctx context.Context, trx repository.Tx) error {
		step = stepAcquireLock
		if err := u.assetImportRepository.AcquireLock(ctx, trx); err != nil {
			return err
		}

		step = stepLoadStoredAssets
		stored, err := u.loadStoredAssets(ctx, trx, file)
		if err != nil {
			return err
		}

		report = assetimport.Validate(file, stored)
		record = u.newImportRecord(input, report)

		// The record goes first: every inserted asset references it.
		step = stepCreateRecord
		if err := u.assetImportRepository.Create(ctx, trx, record); err != nil {
			return err
		}

		if record.Outcome == valueobject.ImportOutcomeCommitted {
			u.stampAssets(report.Accepted, record)
			step = stepCreateAssets
			if err := u.assetRepository.CreateMany(ctx, trx, report.Accepted); err != nil {
				return err
			}
		}

		step = stepCommit
		return nil
	})
	if err != nil {
		log.WithError(err).WithField("step", step).Error("failed to create asset import")
		return usecase.CreateAssetImportOutput{}, err
	}

	log.WithFields(reportLogFields(input.FileName, input.Mode.String(), report)).
		WithFields(logrus.Fields{"importID": record.ID.String(), "outcome": record.Outcome.String()}).
		Info("success")
	return usecase.CreateAssetImportOutput{Import: record, Report: report}, nil
}

// newImportRecord builds the import record of a completed attempt (FR-029).
func (u assetImportUsecase) newImportRecord(input usecase.CreateAssetImportInput, report assetimport.Report) *entity.AssetImport {
	now := u.now()
	return &entity.AssetImport{
		ID:            u.newID(),
		FileName:      input.FileName,
		Mode:          input.Mode,
		Outcome:       importOutcome(input.Mode, report),
		TotalRows:     report.Counts.Total,
		ImportedRows:  report.Counts.ToImport,
		UnchangedRows: report.Counts.Unchanged,
		RejectedRows:  report.Counts.Rejected,
		WarningRows:   report.Counts.WarningRows,
		CreatedAt:     now,
		CreatedByName: usecase.SystemActorName,
		UpdatedAt:     now,
		UpdatedByName: usecase.SystemActorName,
	}
}

// importOutcome decides what a confirm does (data-model.md, commit states).
// Strict mode is checked first: one rejected row blocks the whole file
// (FR-023), even when nothing new would have been saved.
func importOutcome(mode valueobject.ImportMode, report assetimport.Report) valueobject.ImportOutcome {
	if mode == valueobject.ImportModeStrict && report.Counts.Rejected > 0 {
		return valueobject.ImportOutcomeStrictBlocked
	}
	if report.Counts.ToImport == 0 {
		return valueobject.ImportOutcomeNothingToCommit
	}
	return valueobject.ImportOutcomeCommitted
}

// stampAssets gives every accepted asset its identity, its import record and
// the audit columns before insert.
func (u assetImportUsecase) stampAssets(assets []*entity.Asset, record *entity.AssetImport) {
	for _, asset := range assets {
		asset.ID = u.newID()
		asset.ImportID = record.ID
		asset.CreatedAt = record.CreatedAt
		asset.CreatedByName = usecase.SystemActorName
		asset.UpdatedAt = record.UpdatedAt
		asset.UpdatedByName = usecase.SystemActorName
	}
}
