package asset_usecase

import (
	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/usecase"
	"smart-grid-asset-registry/domain/valueobject"
)

// topLevelLogValue names the top level in logs, where there is no parent ID.
const topLevelLogValue = "(top level)"

// parentLogValue is the parentAssetId log field: the ID, or the top level.
func parentLogValue(parentAssetID *string) string {
	if parentAssetID == nil {
		return topLevelLogValue
	}
	return *parentAssetID
}

// toDescendantCounts picks the four counted types; a missing type is 0.
func toDescendantCounts(counts map[valueobject.AssetType]int) *usecase.DescendantCounts {
	return &usecase.DescendantCounts{
		Transformers:      counts[valueobject.AssetTypeTransformer],
		LVBoards:          counts[valueobject.AssetTypeLVBoard],
		Switchboards:      counts[valueobject.AssetTypeSwitchboard],
		SwitchboardPanels: counts[valueobject.AssetTypeSwitchboardPanel],
	}
}

// pageOf returns page number page (from 1) of size limit; a page past the end
// is an empty, non-nil slice. page and limit are already defaulted.
func pageOf(nodes []*entity.AssetNode, page, limit int) []*entity.AssetNode {
	start := (page - 1) * limit
	if start >= len(nodes) {
		return []*entity.AssetNode{}
	}
	return nodes[start:min(start+limit, len(nodes))]
}

// nodeAssetIDs lists the nodes' asset IDs in order.
func nodeAssetIDs(nodes []*entity.AssetNode) []string {
	ids := make([]string, 0, len(nodes))
	for _, node := range nodes {
		ids = append(ids, node.AssetID)
	}
	return ids
}
