package entity

import (
	"slices"
	"strings"
)

// CheckIsExactMatch reports whether the search text equals the asset's whole
// ID or whole name, letter case ignored (spec FR-008).
func CheckIsExactMatch(node *AssetNode, text string) bool {
	return strings.EqualFold(text, node.AssetID) || strings.EqualFold(text, node.AssetName)
}

// SortSearchMatches orders search matches in place: exact matches first, then
// the rest, each part in the explorer's order (SortAssetNodes: type rank,
// then natural asset ID). Because asset IDs are unique, a search for a full ID
// always puts that asset first (research R3).
func SortSearchMatches(nodes []*AssetNode, text string) {
	SortAssetNodes(nodes)
	slices.SortStableFunc(nodes, func(a, b *AssetNode) int {
		aExact, bExact := CheckIsExactMatch(a, text), CheckIsExactMatch(b, text)
		switch {
		case aExact == bExact:
			return 0
		case aExact:
			return -1
		default:
			return 1
		}
	})
}
