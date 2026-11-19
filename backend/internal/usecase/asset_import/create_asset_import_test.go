package asset_import_usecase

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"smart-grid-asset-registry/domain/assetimport"
	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/errorcode"
	"smart-grid-asset-registry/domain/usecase"
	"smart-grid-asset-registry/domain/valueobject"
	"smart-grid-asset-registry/internal/apperr"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// IDs handed out by deps.usecase().newID, in order.
var (
	firstID  = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	secondID = uuid.MustParse("00000000-0000-0000-0000-000000000002")
	thirdID  = uuid.MustParse("00000000-0000-0000-0000-000000000003")
)

// The file: a substation and a transformer are valid, TX-2 has no parent.
var mixedLines = []string{
	"TX-1,SS-1,TRANSFORMER,Tx 1,IN_SERVICE,22,1000,,,,",
	"SS-1,,SUBSTATION,Sub 1,IN_SERVICE,22,,,,,",
	"TX-2,,TRANSFORMER,Tx 2,IN_SERVICE,22,1000,,,,",
}

var mixedIDs = []string{"SS-1", "TX-1", "TX-2"}

func createInput(mode valueobject.ImportMode, lines ...string) usecase.CreateAssetImportInput {
	return usecase.CreateAssetImportInput{FileName: "grid.csv", File: csvFile(lines...), Mode: mode}
}

// expectReads stubs the stored-asset lookups inside the transaction, after the lock.
func expectReads(d deps, lock *mock.Call, ids []string) {
	d.assetRepo.EXPECT().FindByAssetIDs(mock.Anything, d.trx, ids).Return(nil, nil).NotBefore(lock)
	d.assetRepo.EXPECT().FindBySerialNumbers(mock.Anything, d.trx, []string{}).Return(nil, nil).NotBefore(lock)
}

func isStamped(asset *entity.Asset, id uuid.UUID) bool {
	return asset.ID == id && asset.ImportID == firstID &&
		asset.CreatedAt.Equal(fixedNow) && asset.UpdatedAt.Equal(fixedNow) &&
		asset.CreatedByName == usecase.SystemActorName && asset.UpdatedByName == usecase.SystemActorName
}

func TestCreateAssetImport(t *testing.T) {
	ctx := context.Background()
	repoErr := errors.New("db down")

	tests := []struct {
		name    string
		input   usecase.CreateAssetImportInput
		setup   func(d deps)
		check   func(t *testing.T, output usecase.CreateAssetImportOutput)
		wantErr error
	}{
		{
			name:  "partial with new rows: record COMMITTED first, then the stamped assets, then commit",
			input: createInput(valueobject.ImportModePartial, mixedLines...),
			setup: func(d deps) {
				d.withinTx()
				lock := d.assetImportRepo.EXPECT().AcquireLock(mock.Anything, d.trx).Return(nil).Call
				expectReads(d, lock, mixedIDs)
				record := d.assetImportRepo.EXPECT().Create(mock.Anything, d.trx, mock.MatchedBy(func(r *entity.AssetImport) bool {
					return r.ID == firstID && r.FileName == "grid.csv" && r.Mode == valueobject.ImportModePartial &&
						r.Outcome == valueobject.ImportOutcomeCommitted &&
						r.TotalRows == 3 && r.ImportedRows == 2 && r.UnchangedRows == 0 && r.RejectedRows == 1 && r.WarningRows == 0 &&
						r.CreatedAt.Equal(fixedNow) && r.UpdatedAt.Equal(fixedNow) &&
						r.CreatedByName == usecase.SystemActorName && r.UpdatedByName == usecase.SystemActorName
				})).Return(nil).Call
				d.assetRepo.EXPECT().CreateMany(mock.Anything, d.trx, mock.MatchedBy(func(assets []*entity.Asset) bool {
					// parents first: SS-1, then TX-1
					return len(assets) == 2 && assets[0].AssetID == "SS-1" && assets[1].AssetID == "TX-1" &&
						isStamped(assets[0], secondID) && isStamped(assets[1], thirdID)
				})).Return(nil).NotBefore(record)
				d.trx.EXPECT().Commit().Return(nil)
			},
			check: func(t *testing.T, output usecase.CreateAssetImportOutput) {
				require.NotNil(t, output.Import)
				assert.Equal(t, valueobject.ImportOutcomeCommitted, output.Import.Outcome)
				assert.Equal(t, 2, output.Report.Counts.ToImport)
				require.Len(t, output.Report.Rejected, 1)
				assert.Equal(t, errorcode.ParentRequired, output.Report.Rejected[0].Errors[0].Code)
			},
		},
		{
			name:  "no new rows: record NOTHING_TO_COMMIT, no assets inserted, commit",
			input: createInput(valueobject.ImportModePartial, "TX-2,,TRANSFORMER,Tx 2,IN_SERVICE,22,1000,,,,"),
			setup: func(d deps) {
				d.withinTx()
				lock := d.assetImportRepo.EXPECT().AcquireLock(mock.Anything, d.trx).Return(nil).Call
				expectReads(d, lock, []string{"TX-2"})
				d.assetImportRepo.EXPECT().Create(mock.Anything, d.trx, mock.MatchedBy(func(r *entity.AssetImport) bool {
					return r.Outcome == valueobject.ImportOutcomeNothingToCommit && r.ImportedRows == 0 && r.RejectedRows == 1
				})).Return(nil)
				d.trx.EXPECT().Commit().Return(nil)
			},
			check: func(t *testing.T, output usecase.CreateAssetImportOutput) {
				require.NotNil(t, output.Import)
				assert.Equal(t, valueobject.ImportOutcomeNothingToCommit, output.Import.Outcome)
			},
		},
		{
			name:  "strict with a rejected row: record STRICT_BLOCKED, no assets inserted, commit",
			input: createInput(valueobject.ImportModeStrict, mixedLines...),
			setup: func(d deps) {
				d.withinTx()
				lock := d.assetImportRepo.EXPECT().AcquireLock(mock.Anything, d.trx).Return(nil).Call
				expectReads(d, lock, mixedIDs)
				d.assetImportRepo.EXPECT().Create(mock.Anything, d.trx, mock.MatchedBy(func(r *entity.AssetImport) bool {
					// imported_rows still counts the new valid rows; they were not saved
					return r.Mode == valueobject.ImportModeStrict && r.Outcome == valueobject.ImportOutcomeStrictBlocked &&
						r.TotalRows == 3 && r.ImportedRows == 2 && r.RejectedRows == 1
				})).Return(nil)
				// no CreateMany: the mock fails the test if it is called
				d.trx.EXPECT().Commit().Return(nil)
			},
			check: func(t *testing.T, output usecase.CreateAssetImportOutput) {
				require.NotNil(t, output.Import)
				assert.Equal(t, valueobject.ImportOutcomeStrictBlocked, output.Import.Outcome)
				assert.Len(t, output.Report.Rejected, 1, "every rejected row is still reported")
			},
		},
		{
			name:  "strict with no rejected rows: record COMMITTED, assets inserted",
			input: createInput(valueobject.ImportModeStrict, mixedLines[0], mixedLines[1]),
			setup: func(d deps) {
				d.withinTx()
				lock := d.assetImportRepo.EXPECT().AcquireLock(mock.Anything, d.trx).Return(nil).Call
				expectReads(d, lock, []string{"SS-1", "TX-1"})
				d.assetImportRepo.EXPECT().Create(mock.Anything, d.trx, mock.MatchedBy(func(r *entity.AssetImport) bool {
					return r.Mode == valueobject.ImportModeStrict && r.Outcome == valueobject.ImportOutcomeCommitted &&
						r.ImportedRows == 2 && r.RejectedRows == 0
				})).Return(nil)
				d.assetRepo.EXPECT().CreateMany(mock.Anything, d.trx, mock.MatchedBy(func(assets []*entity.Asset) bool {
					return len(assets) == 2
				})).Return(nil)
				d.trx.EXPECT().Commit().Return(nil)
			},
			check: func(t *testing.T, output usecase.CreateAssetImportOutput) {
				require.NotNil(t, output.Import)
				assert.Equal(t, valueobject.ImportOutcomeCommitted, output.Import.Outcome)
			},
		},
		{
			name:  "strict with no rejected and no new rows: record NOTHING_TO_COMMIT, no assets inserted",
			input: createInput(valueobject.ImportModeStrict, "SS-1,,SUBSTATION,Sub 1,IN_SERVICE,22,,,,,"),
			setup: func(d deps) {
				d.withinTx()
				status := valueobject.OperationalStatusInService
				voltage := 22.0
				storedSubstation := &entity.Asset{
					AssetID: "SS-1", AssetType: valueobject.AssetTypeSubstation, AssetName: "Sub 1",
					OperationalStatus: &status, VoltageKV: &voltage,
				}
				lock := d.assetImportRepo.EXPECT().AcquireLock(mock.Anything, d.trx).Return(nil).Call
				d.assetRepo.EXPECT().FindByAssetIDs(mock.Anything, d.trx, []string{"SS-1"}).
					Return([]*entity.Asset{storedSubstation}, nil).NotBefore(lock)
				d.assetRepo.EXPECT().FindBySerialNumbers(mock.Anything, d.trx, []string{}).Return(nil, nil).NotBefore(lock)
				d.assetImportRepo.EXPECT().Create(mock.Anything, d.trx, mock.MatchedBy(func(r *entity.AssetImport) bool {
					return r.Outcome == valueobject.ImportOutcomeNothingToCommit &&
						r.ImportedRows == 0 && r.UnchangedRows == 1 && r.RejectedRows == 0
				})).Return(nil)
				d.trx.EXPECT().Commit().Return(nil)
			},
			check: func(t *testing.T, output usecase.CreateAssetImportOutput) {
				require.NotNil(t, output.Import)
				assert.Equal(t, valueobject.ImportOutcomeNothingToCommit, output.Import.Outcome)
			},
		},
		{
			name:    "invalid request shape: 400, no transaction",
			input:   usecase.CreateAssetImportInput{FileName: "grid.csv", Mode: valueobject.ImportModePartial},
			setup:   func(deps) {},
			wantErr: apperr.BadRequest(errorcode.InvalidRequest, "file is required"),
		},
		{
			name:    "file-level problem: 422, no transaction",
			input:   createInput(valueobject.ImportModePartial),
			setup:   func(deps) {},
			wantErr: apperr.UnprocessableEntity(errorcode.ImportFileEmpty),
		},
		{
			name:  "lock error rolls back",
			input: createInput(valueobject.ImportModePartial, mixedLines...),
			setup: func(d deps) {
				d.withinTx()
				d.assetImportRepo.EXPECT().AcquireLock(mock.Anything, d.trx).Return(repoErr)
				d.trx.EXPECT().Rollback().Return(nil)
			},
			wantErr: repoErr,
		},
		{
			name:  "stored-asset lookup error rolls back",
			input: createInput(valueobject.ImportModePartial, mixedLines...),
			setup: func(d deps) {
				d.withinTx()
				d.assetImportRepo.EXPECT().AcquireLock(mock.Anything, d.trx).Return(nil)
				d.assetRepo.EXPECT().FindByAssetIDs(mock.Anything, d.trx, mixedIDs).Return(nil, repoErr)
				d.trx.EXPECT().Rollback().Return(nil)
			},
			wantErr: repoErr,
		},
		{
			name:  "import record insert error rolls back, no assets inserted",
			input: createInput(valueobject.ImportModePartial, mixedLines...),
			setup: func(d deps) {
				d.withinTx()
				lock := d.assetImportRepo.EXPECT().AcquireLock(mock.Anything, d.trx).Return(nil).Call
				expectReads(d, lock, mixedIDs)
				d.assetImportRepo.EXPECT().Create(mock.Anything, d.trx, mock.Anything).Return(repoErr)
				d.trx.EXPECT().Rollback().Return(nil)
			},
			wantErr: repoErr,
		},
		{
			name:  "asset insert error rolls back",
			input: createInput(valueobject.ImportModePartial, mixedLines...),
			setup: func(d deps) {
				d.withinTx()
				lock := d.assetImportRepo.EXPECT().AcquireLock(mock.Anything, d.trx).Return(nil).Call
				expectReads(d, lock, mixedIDs)
				d.assetImportRepo.EXPECT().Create(mock.Anything, d.trx, mock.Anything).Return(nil)
				d.assetRepo.EXPECT().CreateMany(mock.Anything, d.trx, mock.Anything).Return(repoErr)
				d.trx.EXPECT().Rollback().Return(nil)
			},
			wantErr: repoErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newDeps(t)
			tt.setup(d)

			output, err := d.usecase().CreateAssetImport(ctx, tt.input)

			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
				return
			}
			require.NoError(t, err)
			tt.check(t, output)
		})
	}
}

// TestImportOutcome covers the commit-state decision (data-model.md) on its
// own, including strict files with nothing new and nothing rejected, which
// only arise once rows can be unchanged.
func TestImportOutcome(t *testing.T) {
	counts := func(toImport, unchanged, rejected int) assetimport.Report {
		return assetimport.Report{Counts: assetimport.Counts{
			Total: toImport + unchanged + rejected, ToImport: toImport, Unchanged: unchanged, Rejected: rejected,
		}}
	}

	tests := []struct {
		name   string
		mode   valueobject.ImportMode
		report assetimport.Report
		want   valueobject.ImportOutcome
	}{
		{"partial, new rows", valueobject.ImportModePartial, counts(2, 0, 1), valueobject.ImportOutcomeCommitted},
		{"partial, nothing new", valueobject.ImportModePartial, counts(0, 3, 1), valueobject.ImportOutcomeNothingToCommit},
		{"strict, a rejected row", valueobject.ImportModeStrict, counts(2, 0, 1), valueobject.ImportOutcomeStrictBlocked},
		{"strict, rejected and nothing new", valueobject.ImportModeStrict, counts(0, 0, 1), valueobject.ImportOutcomeStrictBlocked},
		{"strict, clean with new rows", valueobject.ImportModeStrict, counts(2, 1, 0), valueobject.ImportOutcomeCommitted},
		{"strict, clean and nothing new", valueobject.ImportModeStrict, counts(0, 3, 0), valueobject.ImportOutcomeNothingToCommit},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, importOutcome(tt.mode, tt.report))
		})
	}
}

func TestCreateAssetImport_Logging(t *testing.T) {
	ctx := context.Background()

	newLoggingDeps := func(t *testing.T) (deps, assetImportUsecase, *bytes.Buffer) {
		d := newDeps(t)
		var logs bytes.Buffer
		u := d.usecase()
		u.log = logrus.New()
		u.log.SetOutput(&logs)
		u.log.SetFormatter(&logrus.JSONFormatter{})
		return d, u, &logs
	}

	t.Run("success logs the import ID, outcome and counts", func(t *testing.T) {
		d, u, logs := newLoggingDeps(t)
		d.withinTx()
		lock := d.assetImportRepo.EXPECT().AcquireLock(mock.Anything, d.trx).Return(nil).Call
		expectReads(d, lock, mixedIDs)
		d.assetImportRepo.EXPECT().Create(mock.Anything, d.trx, mock.Anything).Return(nil)
		d.assetRepo.EXPECT().CreateMany(mock.Anything, d.trx, mock.Anything).Return(nil)
		d.trx.EXPECT().Commit().Return(nil)

		_, err := u.CreateAssetImport(ctx, createInput(valueobject.ImportModePartial, mixedLines...))

		require.NoError(t, err)
		assert.Contains(t, logs.String(), `"msg":"success"`)
		assert.Contains(t, logs.String(), `"importID":"`+firstID.String()+`"`)
		assert.Contains(t, logs.String(), `"outcome":"COMMITTED"`)
		assert.NotContains(t, logs.String(), "Sub 1", "row contents are never logged")
	})

	t.Run("a technical failure is logged once, with file name, mode and step (FR-030)", func(t *testing.T) {
		d, u, logs := newLoggingDeps(t)
		d.withinTx()
		d.assetImportRepo.EXPECT().AcquireLock(mock.Anything, d.trx).Return(errors.New("db down"))
		d.trx.EXPECT().Rollback().Return(nil)

		_, err := u.CreateAssetImport(ctx, createInput(valueobject.ImportModeStrict, mixedLines...))

		require.Error(t, err)
		assert.Contains(t, logs.String(), `"level":"error"`)
		assert.Contains(t, logs.String(), `"fileName":"grid.csv"`)
		assert.Contains(t, logs.String(), `"mode":"STRICT"`)
		assert.Contains(t, logs.String(), `"step":"acquire import lock"`)
		assert.Equal(t, 1, strings.Count(logs.String(), `"level":"error"`))
	})

	t.Run("a transaction that cannot begin is logged too", func(t *testing.T) {
		d, u, logs := newLoggingDeps(t)
		beginErr := errors.New("failed to connect")
		d.txManager.EXPECT().WithinTx(mock.Anything, mock.Anything).Return(beginErr)

		_, err := u.CreateAssetImport(ctx, createInput(valueobject.ImportModePartial, mixedLines...))

		require.ErrorIs(t, err, beginErr)
		assert.Contains(t, logs.String(), `"step":"begin transaction"`)
		assert.Contains(t, logs.String(), `"fileName":"grid.csv"`)
	})
}
