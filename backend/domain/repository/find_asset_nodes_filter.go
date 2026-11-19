package repository

// FindAssetNodesFilter selects the nodes FindNodes returns. It has a single
// field today on purpose: the search feature will add its filters here
// instead of changing the method signature (data-model.md).
type FindAssetNodesFilter struct {
	// ParentAssetID: nil = the top level (substations, which have no
	// parent); otherwise the immediate children of that asset.
	ParentAssetID *string
}
