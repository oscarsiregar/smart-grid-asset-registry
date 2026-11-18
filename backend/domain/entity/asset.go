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
