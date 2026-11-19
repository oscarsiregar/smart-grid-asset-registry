package entity

import (
	"time"

	"smart-grid-asset-registry/domain/valueobject"

	"github.com/google/uuid"
)

// Asset is the domain representation of a row in the "asset" table: one
// piece of grid equipment, identified by its CSV asset ID. Optional CSV
// values are pointers; nil means the CSV cell was blank.
type Asset struct {
	ID                uuid.UUID
	AssetID           string
	ParentAssetID     *string // nil only for substations
	AssetType         valueobject.AssetType
	AssetName         string
	OperationalStatus *valueobject.OperationalStatus // nil = "no status"
	VoltageKV         *float64
	RatingKVA         *float64
	Manufacturer      *string
	Model             *string
	SerialNumber      *string
	CommissionedDate  *time.Time // date only, UTC midnight
	ImportID          uuid.UUID  // the asset_import that created this asset
	CreatedAt         time.Time
	CreatedByName     string
	UpdatedAt         time.Time
	UpdatedByName     string
}

// DifferingFields returns the CSV column names whose values differ between
// the two assets, in CSV column order; empty when they are the same asset
// data (spec FR-018, FR-019). Numbers compare by value, dates by calendar
// day, and a nil text equals a blank one. ID, ImportID and the audit fields
// are not CSV data and are ignored.
func (a *Asset) DifferingFields(other *Asset) []string {
	comparisons := []struct {
		column string
		equal  bool
	}{
		{"asset_id", a.AssetID == other.AssetID},
		{"parent_asset_id", equalText(a.ParentAssetID, other.ParentAssetID)},
		{"asset_type", a.AssetType == other.AssetType},
		{"asset_name", a.AssetName == other.AssetName},
		{"operational_status", equalStatus(a.OperationalStatus, other.OperationalStatus)},
		{"voltage_kv", equalNumber(a.VoltageKV, other.VoltageKV)},
		{"rating_kva", equalNumber(a.RatingKVA, other.RatingKVA)},
		{"manufacturer", equalText(a.Manufacturer, other.Manufacturer)},
		{"model", equalText(a.Model, other.Model)},
		{"serial_number", equalText(a.SerialNumber, other.SerialNumber)},
		{"commissioned_date", equalDay(a.CommissionedDate, other.CommissionedDate)},
	}

	differing := []string{}
	for _, comparison := range comparisons {
		if !comparison.equal {
			differing = append(differing, comparison.column)
		}
	}
	return differing
}

func equalText(a, b *string) bool {
	text := func(value *string) string {
		if value == nil {
			return ""
		}
		return *value
	}
	return text(a) == text(b)
}

func equalStatus(a, b *valueobject.OperationalStatus) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func equalNumber(a, b *float64) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

// equalDay compares the calendar day each date names in its own location.
func equalDay(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == b
	}
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}
