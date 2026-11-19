package model

// AssetAncestor is a query projection, not a table: one ancestor of an asset
// and how many levels above the asset it is (1 = the parent).
type AssetAncestor struct {
	AssetID    string `gorm:"column:asset_id"`
	AncestorID string `gorm:"column:ancestor_id"`
	Depth      int    `gorm:"column:depth"`
}
