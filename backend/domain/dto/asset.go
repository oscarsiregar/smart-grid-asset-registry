package dto

// The asset explorer HTTP contract (specs/002-asset-explorer/contracts/assets-api.md).
// Lists are complete, not paged (Template Exception 13).

// GetListAssetRequest: an absent or empty ParentAssetID means the top level.
type GetListAssetRequest struct {
	ParentAssetID string `query:"parentAssetId"`
}

// GetAssetDetailRequest: AssetID arrives percent-encoded; the handler decodes
// it (research R2).
type GetAssetDetailRequest struct {
	AssetID string `params:"assetId"`
}

// AssetNodeResponse is the shape of every list item and every path step.
// OperationalStatus is null when no status is recorded.
type AssetNodeResponse struct {
	AssetID           string  `json:"assetId"`
	AssetName         string  `json:"assetName"`
	AssetType         string  `json:"assetType"`
	OperationalStatus *string `json:"operationalStatus"`
	ChildCount        int     `json:"childCount"`
}

// AssetListResponse: Items is never null.
type AssetListResponse struct {
	Items      []AssetNodeResponse `json:"items"`
	TotalItems int                 `json:"totalItems"`
}

// AssetResponse holds every stored attribute; blank CSV values are null and
// CommissionedDate is YYYY-MM-DD.
type AssetResponse struct {
	AssetID           string   `json:"assetId"`
	ParentAssetID     *string  `json:"parentAssetId"`
	AssetType         string   `json:"assetType"`
	AssetName         string   `json:"assetName"`
	OperationalStatus *string  `json:"operationalStatus"`
	VoltageKV         *float64 `json:"voltageKv"`
	RatingKVA         *float64 `json:"ratingKva"`
	Manufacturer      *string  `json:"manufacturer"`
	Model             *string  `json:"model"`
	SerialNumber      *string  `json:"serialNumber"`
	CommissionedDate  *string  `json:"commissionedDate"`
}

// AssetDetailResponse: Path runs from the substation down to the asset;
// Children is never null; DescendantCounts is null unless a substation.
type AssetDetailResponse struct {
	Asset            AssetResponse             `json:"asset"`
	Path             []AssetNodeResponse       `json:"path"`
	Children         []AssetNodeResponse       `json:"children"`
	DescendantCounts *DescendantCountsResponse `json:"descendantCounts"`
}

// DescendantCountsResponse counts every level below a substation by type.
type DescendantCountsResponse struct {
	Transformers      int `json:"transformers"`
	LVBoards          int `json:"lvBoards"`
	Switchboards      int `json:"switchboards"`
	SwitchboardPanels int `json:"switchboardPanels"`
}
