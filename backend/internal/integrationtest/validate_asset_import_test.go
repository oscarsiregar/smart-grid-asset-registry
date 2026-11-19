package integrationtest

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"smart-grid-asset-registry/domain/assetimport"
	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/usecase"
	"smart-grid-asset-registry/domain/valueobject"
	asset_repository "smart-grid-asset-registry/internal/repository/psql/asset"
	"smart-grid-asset-registry/internal/repository/psql/gormtx"
	"smart-grid-asset-registry/internal/testsupport/pgtest"
	asset_import_usecase "smart-grid-asset-registry/internal/usecase/asset_import"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const sampleCSV = "../../domain/assetimport/testdata/grid_assets.csv"

const csvHeader = "asset_id,parent_asset_id,asset_type,asset_name,operational_status," +
	"voltage_kv,rating_kva,manufacturer,model,serial_number,commissioned_date"

func quietLogger() *logrus.Logger {
	log := logrus.New()
	log.SetOutput(io.Discard)
	return log
}

// newValidateUsecase wires real repositories. The import-record repository
// is not needed to validate.
func newValidateUsecase(db *gorm.DB) usecase.AssetImportUsecase {
	return asset_import_usecase.NewAssetImportUsecase(
		gormtx.NewTransactionManager(db), asset_repository.NewAssetRepository(db), nil, quietLogger())
}

func openSample(t *testing.T) io.Reader {
	t.Helper()
	f, err := os.Open(sampleCSV)
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })
	return f
}

func countRows(t *testing.T, db *gorm.DB, table string) int64 {
	t.Helper()
	var count int64
	require.NoError(t, db.Table(table).Count(&count).Error)
	return count
}

func TestValidateAssetImport_SampleOnEmptyStore(t *testing.T) {
	db := pgtest.Open(t)

	output, err := newValidateUsecase(db).ValidateAssetImport(context.Background(), usecase.ValidateAssetImportInput{
		FileName: "grid_assets.csv", File: openSample(t), Mode: valueobject.ImportModePartial,
	})

	require.NoError(t, err)
	assert.Equal(t, assetimport.Counts{Total: 222, ToImport: 205, Unchanged: 0, Rejected: 17, WarningRows: 2}, output.Report.Counts,
		"matches the golden test")
	assert.Zero(t, countRows(t, db, "asset"), "validating stores nothing")
	assert.Zero(t, countRows(t, db, "asset_import"), "validating leaves no import record")
}

func TestValidateAssetImport_ParentAlreadyStored(t *testing.T) {
	db := pgtest.Open(t)
	seedAssets(t, db, seedImport(t, db),
		&entity.Asset{AssetID: "SS-001", AssetType: valueobject.AssetTypeSubstation, AssetName: "Aurora"})

	file := strings.NewReader(csvHeader + "\nTX-NEW,SS-001,TRANSFORMER,New Transformer,IN_SERVICE,22,1000,,,,\n")
	output, err := newValidateUsecase(db).ValidateAssetImport(context.Background(), usecase.ValidateAssetImportInput{
		FileName: "transformer.csv", File: file, Mode: valueobject.ImportModePartial,
	})

	require.NoError(t, err)
	assert.Empty(t, output.Report.Rejected)
	require.Len(t, output.Report.Accepted, 1)
	assert.Equal(t, "TX-NEW", output.Report.Accepted[0].AssetID)
}
