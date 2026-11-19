package assetimport

import (
	"strings"

	"smart-grid-asset-registry/domain/errorcode"
)

// checkReimport compares every row whose own fields are valid with the stored
// asset that has the same asset ID (spec FR-018, FR-019). The same values
// make the row unchanged; any difference rejects it with ASSET_ID_CONFLICT
// naming the differing columns, because stored assets are never edited.
// Rows whose ID is not stored are left alone: they are new.
func checkReimport(candidates []*candidate, stored StoredAssets) {
	for _, c := range candidates {
		if !c.fieldsValid() {
			continue
		}
		storedAsset := stored.ByAssetID[c.assetID]
		if storedAsset == nil {
			continue
		}

		differing := storedAsset.DifferingFields(c.asset)
		if len(differing) == 0 {
			c.unchanged = true
			continue
		}
		c.hierarchyIssues = append(c.hierarchyIssues, Issue{
			Code: errorcode.AssetIDConflict,
			Args: []any{c.assetID, strings.Join(differing, ", ")},
		})
	}
}
