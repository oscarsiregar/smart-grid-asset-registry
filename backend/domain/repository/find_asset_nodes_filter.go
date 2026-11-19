package repository

import "smart-grid-asset-registry/domain/valueobject"

// FindAssetNodesFilter selects the nodes FindNodes returns: the hierarchy of
// feature 002 or, with Search, the matches of feature 003.
type FindAssetNodesFilter struct {
	// ParentAssetID (without Search): nil = the top level (substations,
	// which have no parent); otherwise the immediate children of that asset.
	ParentAssetID *string
	// Search, when set, matches every level by asset ID or name: letter case
	// ignored, every character literal. ParentAssetID must then be nil.
	Search *string
	// AssetType limits a search to one type; only with Search.
	AssetType *valueobject.AssetType
}
