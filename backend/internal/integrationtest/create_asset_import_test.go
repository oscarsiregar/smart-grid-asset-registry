package integrationtest

import (
	"context"
	"errors"
	"strings"
	"testing"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/errorcode"
	"smart-grid-asset-registry/domain/repository"
	"smart-grid-asset-registry/domain/usecase"
	"smart-grid-asset-registry/domain/valueobject"
	asset_repository "smart-grid-asset-registry/internal/repository/psql/asset"
	asset_import_repository "smart-grid-asset-registry/internal/repository/psql/asset_import"
	"smart-grid-asset-registry/internal/repository/psql/gormtx"
	"smart-grid-asset-registry/internal/testsupport/pgtest"
	asset_import_usecase "smart-grid-asset-registry/internal/usecase/asset_import"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newImportUsecase(db *gorm.DB, assets repository.AssetRepository) usecase.AssetImportUsecase {
	return asset_import_usecase.NewAssetImportUsecase(
		gormtx.NewTransactionManager(db), assets, asset_import_repository.NewAssetImportRepository(db), quietLogger())
}

func csvBody(lines ...string) *strings.Reader {
	return strings.NewReader(csvHeader + "\n" + strings.Join(lines, "\n") + "\n")
}

func commit(t *testing.T, uc usecase.AssetImportUsecase, mode valueobject.ImportMode, lines ...string) usecase.CreateAssetImportOutput {
	t.Helper()
	output, err := uc.CreateAssetImport(context.Background(), usecase.CreateAssetImportInput{
		FileName: "file.csv", File: csvBody(lines...), Mode: mode,
	})
	require.NoError(t, err)
	return output
}

// failingAssets is a real repository whose inserts fail, to force a rollback
// in the middle of the transaction.
type failingAssets struct {
	repository.AssetRepository
}

var errInjected = errors.New("injected insert failure")

func (failingAssets) CreateMany(context.Context, repository.Tx, []*entity.Asset) error {
	return errInjected
}

func TestCreateAssetImport_SampleOnEmptyStore(t *testing.T) {
	db := pgtest.Open(t)
	uc := newImportUsecase(db, asset_repository.NewAssetRepository(db))

	output, err := uc.CreateAssetImport(context.Background(), usecase.CreateAssetImportInput{
		FileName: "grid_assets.csv", File: openSample(t), Mode: valueobject.ImportModePartial,
	})

	require.NoError(t, err)
	require.NotNil(t, output.Import)
	assert.Equal(t, valueobject.ImportOutcomeCommitted, output.Import.Outcome)

	assert.EqualValues(t, 205, countRows(t, db, "asset"))
	var substations, wrongAudit, orphans int64
	require.NoError(t, db.Table("asset").Where("asset_type = ?", "SUBSTATION").Count(&substations).Error)
	require.NoError(t, db.Table("asset").Where("created_by_name <> 'system' OR updated_by_name <> 'system'").Count(&wrongAudit).Error)
	require.NoError(t, db.Raw(`SELECT count(*) FROM asset child LEFT JOIN asset parent ON parent.asset_id = child.parent_asset_id
		WHERE child.parent_asset_id IS NOT NULL AND parent.id IS NULL`).Scan(&orphans).Error)
	assert.EqualValues(t, 13, substations)
	assert.Zero(t, wrongAudit)
	assert.Zero(t, orphans, "every parent_asset_id resolves")

	var record struct {
		Outcome                                                           string
		TotalRows, ImportedRows, UnchangedRows, RejectedRows, WarningRows int
	}
	require.NoError(t, db.Table("asset_import").Select("outcome, total_rows, imported_rows, unchanged_rows, rejected_rows, warning_rows").Scan(&record).Error)
	assert.Equal(t, "COMMITTED", record.Outcome)
	assert.Equal(t, []int{222, 205, 0, 17, 2},
		[]int{record.TotalRows, record.ImportedRows, record.UnchangedRows, record.RejectedRows, record.WarningRows})
	assert.EqualValues(t, 1, countRows(t, db, "asset_import"))

	// FR-031: rejected-row and warning details are never stored; there is no
	// table for them.
	var tables []string
	require.NoError(t, db.Raw(`SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' ORDER BY 1`).
		Scan(&tables).Error)
	assert.Equal(t, []string{"asset", "asset_import", "schema_migrations"}, tables)
}

func TestCreateAssetImport_FailureRollsEverythingBack(t *testing.T) {
	db := pgtest.Open(t)
	uc := newImportUsecase(db, failingAssets{AssetRepository: asset_repository.NewAssetRepository(db)})

	_, err := uc.CreateAssetImport(context.Background(), usecase.CreateAssetImportInput{
		FileName: "grid_assets.csv", File: openSample(t), Mode: valueobject.ImportModePartial,
	})

	require.ErrorIs(t, err, errInjected)
	assert.Zero(t, countRows(t, db, "asset"), "no assets (FR-030, SC-006)")
	assert.Zero(t, countRows(t, db, "asset_import"), "no import record for a failed attempt")
}

// TestCreateAssetImport_RevalidatesAtConfirm: a parent that is missing when the
// file is checked but stored before the confirm makes the row valid (FR-022).
func TestCreateAssetImport_RevalidatesAtConfirm(t *testing.T) {
	db := pgtest.Open(t)
	uc := newImportUsecase(db, asset_repository.NewAssetRepository(db))
	transformer := "TX-X,SS-X,TRANSFORMER,Transformer X,IN_SERVICE,22,1000,,,,"

	checked, err := uc.ValidateAssetImport(context.Background(), usecase.ValidateAssetImportInput{
		FileName: "tx.csv", File: csvBody(transformer), Mode: valueobject.ImportModePartial,
	})
	require.NoError(t, err)
	require.Len(t, checked.Report.Rejected, 1)
	assert.Equal(t, errorcode.ParentNotFound, checked.Report.Rejected[0].Errors[0].Code)

	commit(t, uc, valueobject.ImportModePartial, "SS-X,,SUBSTATION,Substation X,IN_SERVICE,22,,,,,")
	confirmed := commit(t, uc, valueobject.ImportModePartial, transformer)

	assert.Equal(t, valueobject.ImportOutcomeCommitted, confirmed.Import.Outcome)
	assert.Equal(t, 1, confirmed.Report.Counts.ToImport)
	assert.EqualValues(t, 2, countRows(t, db, "asset"))
}

// TestCreateAssetImport_ConcurrentImportsAreSerialised runs two imports at the
// same time on non-overlapping files, where file B's transformer needs file
// A's substation. The lock serialises them: whichever runs second sees the
// other's committed assets, so the end state is always consistent (FR-024).
func TestCreateAssetImport_ConcurrentImportsAreSerialised(t *testing.T) {
	db := pgtest.Open(t)
	uc := newImportUsecase(db, asset_repository.NewAssetRepository(db))

	type result struct {
		output usecase.CreateAssetImportOutput
		err    error
	}
	run := func(lines ...string) <-chan result {
		done := make(chan result, 1)
		go func() {
			output, err := uc.CreateAssetImport(context.Background(), usecase.CreateAssetImportInput{
				FileName: "file.csv", File: csvBody(lines...), Mode: valueobject.ImportModePartial,
			})
			done <- result{output, err}
		}()
		return done
	}

	fileA := run("SS-A,,SUBSTATION,Substation A,IN_SERVICE,22,,,,,")
	fileB := run("TX-B,SS-A,TRANSFORMER,Transformer B,IN_SERVICE,22,1000,,,,")
	a, b := <-fileA, <-fileB

	require.NoError(t, a.err)
	require.NoError(t, b.err)
	assert.Equal(t, valueobject.ImportOutcomeCommitted, a.output.Import.Outcome)
	assert.EqualValues(t, 2, countRows(t, db, "asset_import"))

	switch b.output.Import.Outcome {
	case valueobject.ImportOutcomeCommitted: // B ran after A and saw SS-A
		assert.EqualValues(t, 2, countRows(t, db, "asset"))
	case valueobject.ImportOutcomeNothingToCommit: // B ran first: SS-A did not exist yet
		assert.Equal(t, errorcode.ParentNotFound, b.output.Report.Rejected[0].Errors[0].Code)
		assert.EqualValues(t, 1, countRows(t, db, "asset"))
	default:
		t.Fatalf("unexpected outcome for file B: %s", b.output.Import.Outcome)
	}
}
