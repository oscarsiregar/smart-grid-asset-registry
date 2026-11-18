package mapper

import (
	"testing"
	"time"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/valueobject"
	"smart-grid-asset-registry/internal/repository/psql/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func fullAsset() *entity.Asset {
	parent := "SS-001"
	status := valueobject.OperationalStatusInService
	voltage, rating := 22.0, 2000.0
	manufacturer, modelName, serial := "Arcwell Industries", "TXR-2000", "TX00101A"
	commissioned := time.Date(2010, time.March, 12, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, time.November, 19, 10, 0, 0, 0, time.UTC)

	return &entity.Asset{
		ID:                uuid.New(),
		AssetID:           "TX-001-1",
		ParentAssetID:     &parent,
		AssetType:         valueobject.AssetTypeTransformer,
		AssetName:         "Aurora Transformer T1",
		OperationalStatus: &status,
		VoltageKV:         &voltage,
		RatingKVA:         &rating,
		Manufacturer:      &manufacturer,
		Model:             &modelName,
		SerialNumber:      &serial,
		CommissionedDate:  &commissioned,
		ImportID:          uuid.New(),
		CreatedAt:         now,
		CreatedByName:     "system",
		UpdatedAt:         now,
		UpdatedByName:     "system",
	}
}

func TestAssetMapper_RoundTripAllFields(t *testing.T) {
	e := fullAsset()

	m := ToModelAsset(e)
	assert.Equal(t, &model.Asset{
		ID: e.ID, AssetID: "TX-001-1", ParentAssetID: e.ParentAssetID, AssetType: valueobject.AssetTypeTransformer,
		AssetName: "Aurora Transformer T1", OperationalStatus: e.OperationalStatus, VoltageKV: e.VoltageKV,
		RatingKVA: e.RatingKVA, Manufacturer: e.Manufacturer, Model: e.Model, SerialNumber: e.SerialNumber,
		CommissionedDate: e.CommissionedDate, ImportID: e.ImportID, CreatedAt: e.CreatedAt, CreatedByName: "system",
		UpdatedAt: e.UpdatedAt, UpdatedByName: "system",
	}, m)

	assert.Equal(t, e, ToEntityAsset(m))
}

func TestAssetMapper_NilOptionalFields(t *testing.T) {
	e := &entity.Asset{ID: uuid.New(), AssetID: "SS-001", AssetType: valueobject.AssetTypeSubstation, AssetName: "Aurora"}

	m := ToModelAsset(e)
	assert.Nil(t, m.ParentAssetID)
	assert.Nil(t, m.OperationalStatus)
	assert.Nil(t, m.VoltageKV)
	assert.Nil(t, m.RatingKVA)
	assert.Nil(t, m.Manufacturer)
	assert.Nil(t, m.Model)
	assert.Nil(t, m.SerialNumber)
	assert.Nil(t, m.CommissionedDate)
	assert.Equal(t, e, ToEntityAsset(m))
}

func TestAssetMapper_Nil(t *testing.T) {
	assert.Nil(t, ToModelAsset(nil))
	assert.Nil(t, ToEntityAsset(nil))
}

func TestAssetMapper_Slices(t *testing.T) {
	entities := ToEntityAssets([]*model.Asset{{AssetID: "A"}, {AssetID: "B"}})
	assert.Len(t, entities, 2)
	assert.Equal(t, "B", entities[1].AssetID)

	models := ToModelAssets([]*entity.Asset{{AssetID: "A"}})
	assert.Len(t, models, 1)
	assert.Equal(t, "A", models[0].AssetID)

	assert.NotNil(t, ToEntityAssets(nil))
	assert.Empty(t, ToEntityAssets(nil))
	assert.NotNil(t, ToModelAssets(nil))
	assert.Empty(t, ToModelAssets(nil))
}
