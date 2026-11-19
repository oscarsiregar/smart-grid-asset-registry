package asset_usecase

// topLevelLogValue names the top level in logs, where there is no parent ID.
const topLevelLogValue = "(top level)"

// parentLogValue is the parentAssetId log field: the ID, or the top level.
func parentLogValue(parentAssetID *string) string {
	if parentAssetID == nil {
		return topLevelLogValue
	}
	return *parentAssetID
}
