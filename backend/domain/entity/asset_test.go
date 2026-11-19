package entity

import (
	"testing"
	"time"

	"smart-grid-asset-registry/domain/valueobject"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func pointer[T any](value T) *T { return &value }

// fullAsset has every one of the eleven CSV columns filled.
func fullAsset() *Asset {
	return &Asset{
		AssetID:           "TX-1",
		ParentAssetID:     pointer("SS-1"),
		AssetType:         valueobject.AssetTypeTransformer,
		AssetName:         "Transformer 1",
		OperationalStatus: pointer(valueobject.OperationalStatusInService),
		VoltageKV:         pointer(22.0),
		RatingKVA:         pointer(1000.0),
		Manufacturer:      pointer("Arcwell"),
		Model:             pointer("TXR-2000"),
		SerialNumber:      pointer("TX001"),
		CommissionedDate:  pointer(time.Date(2010, 3, 12, 0, 0, 0, 0, time.UTC)),
	}
}

func TestAsset_DifferingFields(t *testing.T) {
	tests := []struct {
		name   string
		change func(a *Asset)
		want   []string
	}{
		{"identical", func(*Asset) {}, []string{}},
		{"number written differently but equal by value (22 vs 22.0)", func(a *Asset) {
			a.VoltageKV = pointer(22.000)
		}, []string{}},
		{"same calendar day at another time and zone", func(a *Asset) {
			a.CommissionedDate = pointer(time.Date(2010, 3, 12, 15, 30, 0, 0, time.FixedZone("WIB", 7*60*60)))
		}, []string{}},
		{"blank text differs from a value", func(a *Asset) {
			a.Manufacturer = pointer("")
		}, []string{"manufacturer"}},
		{"audit fields, ID and ImportID are ignored", func(a *Asset) {
			a.ID = uuid.New()
			a.ImportID = uuid.New()
			a.CreatedAt = time.Now()
			a.CreatedByName = "someone"
			a.UpdatedAt = time.Now()
			a.UpdatedByName = "someone"
		}, []string{}},
		{"asset_id", func(a *Asset) { a.AssetID = "TX-2" }, []string{"asset_id"}},
		{"parent_asset_id", func(a *Asset) { a.ParentAssetID = pointer("SS-2") }, []string{"parent_asset_id"}},
		{"parent_asset_id removed", func(a *Asset) { a.ParentAssetID = nil }, []string{"parent_asset_id"}},
		{"asset_type", func(a *Asset) { a.AssetType = valueobject.AssetTypeLVBoard }, []string{"asset_type"}},
		{"asset_name", func(a *Asset) { a.AssetName = "Other" }, []string{"asset_name"}},
		{"operational_status", func(a *Asset) {
			a.OperationalStatus = pointer(valueobject.OperationalStatusMaintenance)
		}, []string{"operational_status"}},
		{"operational_status removed", func(a *Asset) { a.OperationalStatus = nil }, []string{"operational_status"}},
		{"voltage_kv", func(a *Asset) { a.VoltageKV = pointer(11.0) }, []string{"voltage_kv"}},
		{"rating_kva", func(a *Asset) { a.RatingKVA = nil }, []string{"rating_kva"}},
		{"manufacturer", func(a *Asset) { a.Manufacturer = pointer("Other") }, []string{"manufacturer"}},
		{"model", func(a *Asset) { a.Model = pointer("Other") }, []string{"model"}},
		{"serial_number", func(a *Asset) { a.SerialNumber = pointer("Other") }, []string{"serial_number"}},
		{"commissioned_date", func(a *Asset) {
			a.CommissionedDate = pointer(time.Date(2010, 3, 13, 0, 0, 0, 0, time.UTC))
		}, []string{"commissioned_date"}},
		{"several fields, in CSV column order", func(a *Asset) {
			a.CommissionedDate = nil
			a.OperationalStatus = nil
			a.AssetName = "Other"
		}, []string{"asset_name", "operational_status", "commissioned_date"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			other := fullAsset()
			tt.change(other)

			assert.Equal(t, tt.want, fullAsset().DifferingFields(other))
		})
	}

	t.Run("nil and blank text are equal both ways", func(t *testing.T) {
		a, b := fullAsset(), fullAsset()
		a.Model, b.Model = nil, pointer("")
		assert.Empty(t, a.DifferingFields(b))
		assert.Empty(t, b.DifferingFields(a))
	})
}
