package model

import "smart-grid-asset-registry/domain/valueobject"

// AssetNode is a query projection, not a table: the columns of an asset that
// a list or the tree shows, plus its number of immediate children.
type AssetNode struct {
	AssetID           string                         `gorm:"column:asset_id"`
	AssetName         string                         `gorm:"column:asset_name"`
	AssetType         valueobject.AssetType          `gorm:"column:asset_type"`
	OperationalStatus *valueobject.OperationalStatus `gorm:"column:operational_status"`
	ChildCount        int                            `gorm:"column:child_count"`
}
