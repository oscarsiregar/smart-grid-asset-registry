package model

import "smart-grid-asset-registry/domain/valueobject"

// AssetTypeCount is a query projection, not a table: how many assets of one
// type a query counted.
type AssetTypeCount struct {
	AssetType  valueobject.AssetType `gorm:"column:asset_type"`
	AssetCount int                   `gorm:"column:asset_count"`
}
