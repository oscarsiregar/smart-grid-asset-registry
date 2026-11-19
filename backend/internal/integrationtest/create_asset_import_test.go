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

// TestCreateAssetImport_StrictSampleIsBlocked: in strict mode the sample's 17
// rejected rows block the whole file; only the attempt is recorded (FR-023,
// SC-006).
func TestCreateAssetImport_StrictSampleIsBlocked(t *testing.T) {
	db := pgtest.Open(t)
	uc := newImportUsecase(db, asset_repository.NewAssetRepository(db))

	output, err := uc.CreateAssetImport(context.Background(), usecase.CreateAssetImportInput{
		FileName: "grid_assets.csv", File: openSample(t), Mode: valueobject.ImportModeStrict,
	})

	require.NoError(t, err)
	require.NotNil(t, output.Import)
	assert.Equal(t, valueobject.ImportOutcomeStrictBlocked, output.Import.Outcome)
	assert.Len(t, output.Report.Rejected, 17, "every rejected row is still reported")
	assert.Zero(t, countRows(t, db, "asset"))

	var record struct {
		Mode, Outcome                                                     string
		TotalRows, ImportedRows, UnchangedRows, RejectedRows, WarningRows int
	}
	require.NoError(t, db.Table("asset_import").
		Select("mode, outcome, total_rows, imported_rows, unchanged_rows, rejected_rows, warning_rows").Scan(&record).Error)
	assert.Equal(t, "STRICT", record.Mode)
	assert.Equal(t, "STRICT_BLOCKED", record.Outcome)
	assert.Equal(t, []int{222, 205, 0, 17, 2},
		[]int{record.TotalRows, record.ImportedRows, record.UnchangedRows, record.RejectedRows, record.WarningRows})
	assert.EqualValues(t, 1, countRows(t, db, "asset_import"))
}

func TestCreateAssetImport_StrictCleanFileCommits(t *testing.T) {
	db := pgtest.Open(t)
	uc := newImportUsecase(db, asset_repository.NewAssetRepository(db))

	output := commit(t, uc, valueobject.ImportModeStrict,
		"SS-C,,SUBSTATION,Substation C,IN_SERVICE,22,,,,,",
		"TX-C,SS-C,TRANSFORMER,Transformer C,IN_SERVICE,22,1000,,,,",
		"LV-C,SS-C,LV_BOARD,Board C,IN_SERVICE,0.4,,,,,",
	)

	assert.Equal(t, valueobject.ImportOutcomeCommitted, output.Import.Outcome)
	assert.EqualValues(t, 3, countRows(t, db, "asset"))
}

func importSample(t *testing.T, uc usecase.AssetImportUsecase) usecase.CreateAssetImportOutput {
	t.Helper()
	output, err := uc.CreateAssetImport(context.Background(), usecase.CreateAssetImportInput{
		FileName: "grid_assets.csv", File: openSample(t), Mode: valueobject.ImportModePartial,
	})
	require.NoError(t, err)
	return output
}

// TestCreateAssetImport_Reimport runs the re-import rules against values read
// back from PostgreSQL (dates, NUMERIC, NULLs), not just in memory (US4).
func TestCreateAssetImport_Reimport(t *testing.T) {
	db := pgtest.Open(t)
	uc := newImportUsecase(db, asset_repository.NewAssetRepository(db))
	first := importSample(t, uc)
	require.Equal(t, valueobject.ImportOutcomeCommitted, first.Import.Outcome)

	t.Run("the sample a second time adds nothing (FR-018)", func(t *testing.T) {
		second := importSample(t, uc)

		assert.Equal(t, valueobject.ImportOutcomeNothingToCommit, second.Import.Outcome)
		assert.Equal(t, []int{222, 0, 205, 17},
			[]int{second.Import.TotalRows, second.Import.ImportedRows, second.Import.UnchangedRows, second.Import.RejectedRows})
		assert.Equal(t, first.Report.Rejected, second.Report.Rejected, "the same 17 rejections")
		assert.EqualValues(t, 205, countRows(t, db, "asset"))
		assert.EqualValues(t, 2, countRows(t, db, "asset_import"))
	})

	t.Run("a stored ID with a changed status is ASSET_ID_CONFLICT naming the field (FR-019)", func(t *testing.T) {
		output := commit(t, uc, valueobject.ImportModePartial,
			"TX-001-1,SS-001,TRANSFORMER,Aurora Transformer T1,MAINTENANCE,22,2000,Arcwell Industries,TXR-2000,TX00101A,12/3/10")

		assert.Equal(t, valueobject.ImportOutcomeNothingToCommit, output.Import.Outcome)
		require.Len(t, output.Report.Rejected, 1)
		assert.Equal(t, errorcode.AssetIDConflict, output.Report.Rejected[0].Errors[0].Code)
		assert.Equal(t, []any{"TX-001-1", "operational_status"}, output.Report.Rejected[0].Errors[0].Args)

		var status string
		require.NoError(t, db.Table("asset").Select("operational_status").Where("asset_id = ?", "TX-001-1").Scan(&status).Error)
		assert.Equal(t, "IN_SERVICE", status, "the stored asset is not edited")
	})

	t.Run("a transformer-only file under stored SS-001 commits", func(t *testing.T) {
		output := commit(t, uc, valueobject.ImportModePartial,
			"TX-001-9,SS-001,TRANSFORMER,Aurora Transformer T9,IN_SERVICE,22,500,,,,")

		assert.Equal(t, valueobject.ImportOutcomeCommitted, output.Import.Outcome)
		assert.EqualValues(t, 206, countRows(t, db, "asset"))
	})
}

// TestCreateAssetImport_ConflictStoredBetweenValidateAndConfirm: a row that was
// new when checked is rejected at confirm when another import has meanwhile
// stored the same ID with different values (US2 scenario 3, FR-022).
func TestCreateAssetImport_ConflictStoredBetweenValidateAndConfirm(t *testing.T) {
	db := pgtest.Open(t)
	uc := newImportUsecase(db, asset_repository.NewAssetRepository(db))
	ours := "SS-Q,,SUBSTATION,Substation Q,IN_SERVICE,22,,,,,"

	checked, err := uc.ValidateAssetImport(context.Background(), usecase.ValidateAssetImportInput{
		FileName: "q.csv", File: csvBody(ours), Mode: valueobject.ImportModePartial,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, checked.Report.Counts.ToImport)

	commit(t, uc, valueobject.ImportModePartial, "SS-Q,,SUBSTATION,Other name,IN_SERVICE,22,,,,,")
	confirmed := commit(t, uc, valueobject.ImportModePartial, ours)

	assert.Equal(t, valueobject.ImportOutcomeNothingToCommit, confirmed.Import.Outcome)
	require.Len(t, confirmed.Report.Rejected, 1)
	assert.Equal(t, errorcode.AssetIDConflict, confirmed.Report.Rejected[0].Errors[0].Code)
	assert.EqualValues(t, 1, countRows(t, db, "asset"))
}

// TestCreateAssetImport_ConcurrentOverlappingImports: two imports sharing rows
// run at the same time. The lock serialises them, so the second sees the
// first's rows as unchanged instead of failing on the unique key (FR-024).
func TestCreateAssetImport_ConcurrentOverlappingImports(t *testing.T) {
	db := pgtest.Open(t)
	uc := newImportUsecase(db, asset_repository.NewAssetRepository(db))
	shared := []string{
		"SS-O,,SUBSTATION,Substation O,IN_SERVICE,22,,,,,",
		"TX-O,SS-O,TRANSFORMER,Transformer O,IN_SERVICE,22,1000,,,,",
	}

	type result struct {
		output usecase.CreateAssetImportOutput
		err    error
	}
	run := func(extra string) <-chan result {
		done := make(chan result, 1)
		go func() {
			output, err := uc.CreateAssetImport(context.Background(), usecase.CreateAssetImportInput{
				FileName: "file.csv", File: csvBody(append(append([]string(nil), shared...), extra)...),
				Mode: valueobject.ImportModePartial,
			})
			done <- result{output, err}
		}()
		return done
	}

	fileA := run("LV-A,SS-O,LV_BOARD,Board A,IN_SERVICE,0.4,,,,,")
	fileB := run("LV-B,SS-O,LV_BOARD,Board B,IN_SERVICE,0.4,,,,,")
	a, b := <-fileA, <-fileB

	require.NoError(t, a.err)
	require.NoError(t, b.err)
	unchanged := []int{a.output.Report.Counts.Unchanged, b.output.Report.Counts.Unchanged}
	assert.ElementsMatch(t, []int{0, 2}, unchanged, "whichever ran second saw the shared rows as unchanged")
	assert.Equal(t, 4, a.output.Report.Counts.ToImport+b.output.Report.Counts.ToImport, "3 new, then 1 new")
	assert.EqualValues(t, 4, countRows(t, db, "asset"))
	assert.EqualValues(t, 2, countRows(t, db, "asset_import"))
}
