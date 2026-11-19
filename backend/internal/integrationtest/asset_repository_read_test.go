package integrationtest

import (
	"context"
	"testing"
	"time"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/valueobject"
	asset_repository "smart-grid-asset-registry/internal/repository/psql/asset"
	"smart-grid-asset-registry/internal/testsupport/pgtest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssetRepository_ReadsAgainstPostgreSQL(t *testing.T) {
	db := pgtest.Open(t)
	ctx := context.Background()
	importID := seedImport(t, db)

	status := valueobject.OperationalStatusMaintenance
	commissioned := time.Date(2009, time.February, 15, 0, 0, 0, 0, time.UTC)
	seedAssets(t, db, importID,
		&entity.Asset{AssetID: "SS-1", AssetType: valueobject.AssetTypeSubstation, AssetName: "Sub, \"North\" {1}"},
		&entity.Asset{
			AssetID: "TX-1", ParentAssetID: pointer("SS-1"), AssetType: valueobject.AssetTypeTransformer,
			AssetName: "Transformer 1", OperationalStatus: &status, VoltageKV: pointer(0.4), RatingKVA: pointer(2000.0),
			Manufacturer: pointer("Arcwell"), Model: pointer("TXR"), SerialNumber: pointer("SER-1"),
			CommissionedDate: &commissioned,
		},
		&entity.Asset{AssetID: "TX-2", ParentAssetID: pointer("SS-1"), AssetType: valueobject.AssetTypeTransformer,
			AssetName: "Transformer 2", SerialNumber: pointer("SER-1")},
	)
	repository := asset_repository.NewAssetRepository(db)

	t.Run("FindByAssetIDs returns only the requested assets, with every field", func(t *testing.T) {
		assets, err := repository.FindByAssetIDs(ctx, nil, []string{"TX-1", "SS-1", "MISSING"})

		require.NoError(t, err)
		require.Len(t, assets, 2)
		assert.Equal(t, "SS-1", assets[0].AssetID)
		assert.Nil(t, assets[0].ParentAssetID)
		assert.Nil(t, assets[0].OperationalStatus)

		tx := assets[1]
		assert.Equal(t, "TX-1", tx.AssetID)
		assert.Equal(t, "SS-1", *tx.ParentAssetID)
		assert.Equal(t, valueobject.OperationalStatusMaintenance, *tx.OperationalStatus)
		assert.InDelta(t, 0.4, *tx.VoltageKV, 1e-9)
		assert.InDelta(t, 2000.0, *tx.RatingKVA, 1e-9)
		assert.Equal(t, "SER-1", *tx.SerialNumber)
		assert.True(t, commissioned.Equal(tx.CommissionedDate.UTC()), "got %v", tx.CommissionedDate)
		assert.Equal(t, importID, tx.ImportID)
	})

	t.Run("FindByAssetIDs handles values with quotes, commas and braces", func(t *testing.T) {
		assets, err := repository.FindByAssetIDs(ctx, nil, []string{"SS-1", "a \"quoted\", {odd} id"})

		require.NoError(t, err)
		assert.Len(t, assets, 1)
	})

	t.Run("FindBySerialNumbers returns every asset sharing a serial", func(t *testing.T) {
		assets, err := repository.FindBySerialNumbers(ctx, nil, []string{"SER-1"})

		require.NoError(t, err)
		require.Len(t, assets, 2)
		assert.Equal(t, "TX-1", assets[0].AssetID)
		assert.Equal(t, "TX-2", assets[1].AssetID)
	})

	t.Run("empty lists return empty results", func(t *testing.T) {
		byID, err := repository.FindByAssetIDs(ctx, nil, nil)
		require.NoError(t, err)
		assert.Empty(t, byID)

		bySerial, err := repository.FindBySerialNumbers(ctx, nil, []string{"NONE"})
		require.NoError(t, err)
		assert.Empty(t, bySerial)
	})
}
