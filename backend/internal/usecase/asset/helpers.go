package asset_usecase

import (
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
