package asset_import_usecase

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/errorcode"
	"smart-grid-asset-registry/domain/usecase"
	"smart-grid-asset-registry/domain/valueobject"
	"smart-grid-asset-registry/internal/apperr"
	mocksrepo "smart-grid-asset-registry/internal/mocks/repository"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fixedNow = time.Date(2026, 11, 19, 10, 0, 0, 0, time.UTC)

const testHeader = "asset_id,parent_asset_id,asset_type,asset_name,operational_status," +
	"voltage_kv,rating_kva,manufacturer,model,serial_number,commissioned_date"

func csvFile(lines ...string) io.Reader {
	return strings.NewReader(testHeader + "\n" + strings.Join(lines, "\n") + "\n")
}

func newTestLogger() *logrus.Logger {
	log := logrus.New()
	log.SetOutput(io.Discard)
	return log
}

// deps bundles the mocks one test case needs. Mocks are created per case so
// mockery's AssertExpectations (registered via t.Cleanup) checks exactly the
// calls that case expects — an unexpected call fails the test.
type deps struct {
	txManager       *mocksrepo.TransactionManager
	trx             *mocksrepo.Tx
	assetRepo       *mocksrepo.AssetRepository
	assetImportRepo *mocksrepo.AssetImportRepository
}

func newDeps(t *testing.T) deps {
	return deps{
		txManager:       mocksrepo.NewTransactionManager(t),
		trx:             mocksrepo.NewTx(t),
		assetRepo:       mocksrepo.NewAssetRepository(t),
		assetImportRepo: mocksrepo.NewAssetImportRepository(t),
	}
}

// withinTx wires txManager.WithinTx to run the callback against d.trx.
func (d deps) withinTx() {
	mocksrepo.StubWithinTx(d.txManager, d.trx)
}

func (d deps) usecase() assetImportUsecase {
	ids := []uuid.UUID{
		uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		uuid.MustParse("00000000-0000-0000-0000-000000000002"),
		uuid.MustParse("00000000-0000-0000-0000-000000000003"),
		uuid.MustParse("00000000-0000-0000-0000-000000000004"),
	}
	next := 0
	return assetImportUsecase{
		txManager:             d.txManager,
		assetRepository:       d.assetRepo,
		assetImportRepository: d.assetImportRepo,
		log:                   newTestLogger(),
		now:                   func() time.Time { return fixedNow },
		newID: func() uuid.UUID {
			id := ids[next%len(ids)]
			next++
			return id
		},
	}
}

func TestNewAssetImportUsecase(t *testing.T) {
	assert.NotNil(t, NewAssetImportUsecase(nil, nil, nil, newTestLogger()))
}

func TestParseInput(t *testing.T) {
	valid := usecase.ValidateAssetImportInput{FileName: "a.csv", Mode: valueobject.ImportModePartial}

	tests := []struct {
		name    string
		input   usecase.ValidateAssetImportInput
		file    io.Reader
		wantErr error
	}{
		{
			name:    "invalid request shape is 400 INVALID_REQUEST",
			input:   usecase.ValidateAssetImportInput{FileName: "a.csv", Mode: "ALL"},
			file:    csvFile("SS-1,,SUBSTATION,Sub,,,,,,,"),
			wantErr: apperr.BadRequest(errorcode.InvalidRequest, "invalid import mode"),
		},
		{
			name:    "empty file is 422 IMPORT_FILE_EMPTY",
			input:   valid,
			file:    strings.NewReader(""),
			wantErr: apperr.UnprocessableEntity(errorcode.ImportFileEmpty),
		},
		{
			name:    "CSV syntax error is 422 IMPORT_FILE_NOT_CSV with the line",
			input:   valid,
			file:    csvFile("SS-1,,SUBSTATION,\"open,,,,,,,"),
			wantErr: apperr.UnprocessableEntity(errorcode.ImportFileNotCSV, 2),
		},
		{
			name:    "repeated header column is 422 IMPORT_DUPLICATE_COLUMNS",
			input:   valid,
			file:    strings.NewReader(testHeader + ",asset_id\nSS-1,,SUBSTATION,Sub,,,,,,,,SS-1\n"),
			wantErr: apperr.UnprocessableEntity(errorcode.ImportDuplicateColumns, "asset_id"),
		},
		{
			name:    "file over 5 MB is 422 IMPORT_FILE_TOO_LARGE",
			input:   valid,
			file:    strings.NewReader(strings.Repeat("x", usecase.MaxFileBytes+1)),
			wantErr: apperr.UnprocessableEntity(errorcode.ImportFileTooLarge, "5 MB"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.input.File = tt.file

			_, err := parseInput(tt.input, tt.input.File)

			assert.Equal(t, tt.wantErr, err)
		})
	}

	t.Run("a valid file is parsed", func(t *testing.T) {
		input := valid
		input.File = csvFile("SS-1,,SUBSTATION,Sub,,,,,,,")

		file, err := parseInput(input, input.File)

		require.NoError(t, err)
		assert.Len(t, file.Rows, 1)
	})
}

func TestLoadStoredAssets(t *testing.T) {
	ctx := context.Background()
	input := usecase.ValidateAssetImportInput{
		FileName: "a.csv",
		Mode:     valueobject.ImportModePartial,
		File:     csvFile("TX-1,SS-1,TRANSFORMER,Tx,,,,,,SER-1,", "TX-2,SS-1,TRANSFORMER,Tx,,,,,,,"),
	}
	file, err := parseInput(input, input.File)
	require.NoError(t, err)

	serial := "SER-1"
	storedSubstation := &entity.Asset{AssetID: "SS-1", AssetType: valueobject.AssetTypeSubstation}
	storedSibling := &entity.Asset{AssetID: "TX-9", AssetType: valueobject.AssetTypeTransformer, SerialNumber: &serial}

	t.Run("asks once for the referenced IDs and once for the serials", func(t *testing.T) {
		d := newDeps(t)
		d.assetRepo.EXPECT().FindByAssetIDs(ctx, d.trx, []string{"SS-1", "TX-1", "TX-2"}).
			Return([]*entity.Asset{storedSubstation}, nil).Once()
		d.assetRepo.EXPECT().FindBySerialNumbers(ctx, d.trx, []string{"SER-1"}).
			Return([]*entity.Asset{storedSibling}, nil).Once()

		stored, err := d.usecase().loadStoredAssets(ctx, d.trx, file)

		require.NoError(t, err)
		assert.Equal(t, map[string]*entity.Asset{"SS-1": storedSubstation}, stored.ByAssetID)
		assert.Equal(t, map[string][]string{"SER-1": {"TX-9"}}, stored.BySerial)
	})

	t.Run("returns a FindByAssetIDs error", func(t *testing.T) {
		d := newDeps(t)
		repoErr := errors.New("db down")
		d.assetRepo.EXPECT().FindByAssetIDs(ctx, nil, []string{"SS-1", "TX-1", "TX-2"}).Return(nil, repoErr)

		_, err := d.usecase().loadStoredAssets(ctx, nil, file)

		assert.ErrorIs(t, err, repoErr)
	})

	t.Run("returns a FindBySerialNumbers error", func(t *testing.T) {
		d := newDeps(t)
		repoErr := errors.New("db down")
		d.assetRepo.EXPECT().FindByAssetIDs(ctx, nil, []string{"SS-1", "TX-1", "TX-2"}).Return(nil, nil)
		d.assetRepo.EXPECT().FindBySerialNumbers(ctx, nil, []string{"SER-1"}).Return(nil, repoErr)

		_, err := d.usecase().loadStoredAssets(ctx, nil, file)

		assert.ErrorIs(t, err, repoErr)
	})
}
