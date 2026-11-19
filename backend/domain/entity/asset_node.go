package entity

import (
	"cmp"
	"slices"
	"strings"
	"unicode/utf8"

	"smart-grid-asset-registry/domain/valueobject"
)

// AssetNode is an asset as it appears in a list or the tree: identity, label
// fields and how many immediate children it has.
type AssetNode struct {
	AssetID           string
	AssetName         string
	AssetType         valueobject.AssetType
	OperationalStatus *valueobject.OperationalStatus // nil = no status recorded
	ChildCount        int                            // > 0 makes the node expandable (spec FR-003)
}

// SortAssetNodes orders nodes in place by type rank, then by asset ID in
// natural order (CompareAssetIDs). The order is total, so equal input always
// gives the same output (spec FR-004, research R3).
func SortAssetNodes(nodes []*AssetNode) {
	slices.SortStableFunc(nodes, func(a, b *AssetNode) int {
		if byType := cmp.Compare(a.AssetType.Rank(), b.AssetType.Rank()); byType != 0 {
			return byType
		}
		return CompareAssetIDs(a.AssetID, b.AssetID)
	})
}

// CompareAssetIDs compares two asset IDs naturally: runs of digits compare as
// numbers (SS-2 < SS-10), other characters by Unicode code point, and a digit
// run comes before a non-digit character (SS-001 < SS-T01). Leading zeros are
// ignored; when they are the only difference, the first run with fewer zeros
// comes first (SS-1 < SS-01). Different strings never compare equal.
func CompareAssetIDs(a, b string) int {
	zeroTie := 0
	for a != "" && b != "" {
		aDigits, bDigits := digitRun(a), digitRun(b)

		switch {
		case aDigits != "" && bDigits != "":
			if byValue := compareNumbers(aDigits, bDigits); byValue != 0 {
				return byValue
			}
			if zeroTie == 0 {
				zeroTie = cmp.Compare(len(aDigits), len(bDigits))
			}
			a, b = a[len(aDigits):], b[len(bDigits):]
		case aDigits != "":
			return -1
		case bDigits != "":
			return 1
		default:
			aRune, aSize := utf8.DecodeRuneInString(a)
			bRune, bSize := utf8.DecodeRuneInString(b)
			if aRune != bRune {
				return cmp.Compare(aRune, bRune)
			}
			a, b = a[aSize:], b[bSize:]
		}
	}

	if byLength := cmp.Compare(len(a), len(b)); byLength != 0 {
		return byLength // the one that ran out first is a prefix and comes first
	}
	return zeroTie
}

// digitRun returns the ASCII digits at the start of s.
func digitRun(s string) string {
	end := strings.IndexFunc(s, func(r rune) bool { return r < '0' || r > '9' })
	if end < 0 {
		return s
	}
	return s[:end]
}

// compareNumbers compares two digit strings by value, of any length.
func compareNumbers(a, b string) int {
	a, b = strings.TrimLeft(a, "0"), strings.TrimLeft(b, "0")
	if byLength := cmp.Compare(len(a), len(b)); byLength != 0 {
		return byLength
	}
	return strings.Compare(a, b)
}
