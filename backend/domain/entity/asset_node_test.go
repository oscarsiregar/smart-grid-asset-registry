package entity

import (
	"math/rand"
	"testing"

	"smart-grid-asset-registry/domain/valueobject"

	"github.com/stretchr/testify/assert"
)

// sign reduces a comparison result to -1, 0 or 1.
func sign(value int) int {
	switch {
	case value < 0:
		return -1
	case value > 0:
		return 1
	default:
		return 0
	}
}

func TestCompareAssetIDs(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want int // -1: a first, 0: equal, 1: b first
	}{
		{"digit runs compare as numbers", "SS-2", "SS-10", -1},
		{"digit runs compare as numbers, reversed", "SS-10", "SS-2", 1},
		{"zero-padded numbers", "SS-001", "SS-012", -1},
		{"a digit run comes before a non-digit run", "SS-012", "SS-T01", -1},
		{"a digit run comes before a non-digit run, reversed", "SS-T01", "SS-001", 1},
		{"several digit runs", "PNL-001-1-2", "PNL-001-1-10", -1},
		{"equal IDs", "SS-001", "SS-001", 0},
		{"leading zeros ignored, then the shorter run first", "SS-1", "SS-01", -1},
		{"leading zeros: a later difference wins over the zero tie", "SS-01-1", "SS-1-2", -1},
		{"leading zeros only break a tie", "SS-01", "SS-2", -1},
		{"non-digit runs by code point", "SS-A", "SS-B", -1},
		{"upper case before lower case (code point)", "SS-Z", "SS-a", -1},
		{"a prefix comes first", "SS", "SS-1", -1},
		{"empty before anything", "", "A", -1},
		{"numbers larger than an int64", "X-99999999999999999999", "X-100000000000000000000", -1},
		{"non-ASCII by code point", "SS-Å", "SS-Ö", -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, sign(CompareAssetIDs(tt.a, tt.b)))
			assert.Equal(t, -tt.want, sign(CompareAssetIDs(tt.b, tt.a)), "antisymmetric")
		})
	}
}

func TestCompareAssetIDs_NeverZeroForDifferentStrings(t *testing.T) {
	pairs := [][2]string{{"SS-01", "SS-1"}, {"A0", "A00"}, {"0", "00"}, {"x007y", "x7y"}}
	for _, pair := range pairs {
		assert.NotZerof(t, CompareAssetIDs(pair[0], pair[1]), "%q vs %q", pair[0], pair[1])
	}
}

func node(assetID string, assetType valueobject.AssetType) *AssetNode {
	return &AssetNode{AssetID: assetID, AssetName: "Name " + assetID, AssetType: assetType}
}

func assetIDs(nodes []*AssetNode) []string {
	ids := make([]string, 0, len(nodes))
	for _, n := range nodes {
		ids = append(ids, n.AssetID)
	}
	return ids
}

func TestSortAssetNodes_OrdersByTypeThenAssetID(t *testing.T) {
	nodes := []*AssetNode{
		node("SWB-001-2", valueobject.AssetTypeSwitchboard),
		node("PNL-001-1-10", valueobject.AssetTypeSwitchboardPanel),
		node("TX-001-10", valueobject.AssetTypeTransformer),
		node("LVB-001-1", valueobject.AssetTypeLVBoard),
		node("SWB-001-1", valueobject.AssetTypeSwitchboard),
		node("TX-001-2", valueobject.AssetTypeTransformer),
		node("PNL-001-1-2", valueobject.AssetTypeSwitchboardPanel),
	}

	SortAssetNodes(nodes)

	assert.Equal(t, []string{
		"TX-001-2", "TX-001-10", "LVB-001-1", "SWB-001-1", "SWB-001-2", "PNL-001-1-2", "PNL-001-1-10",
	}, assetIDs(nodes))
}

func TestSortAssetNodes_SampleSubstations(t *testing.T) {
	nodes := []*AssetNode{}
	for _, id := range []string{"SS-T01", "SS-012", "SS-002", "SS-010", "SS-001"} {
		nodes = append(nodes, node(id, valueobject.AssetTypeSubstation))
	}

	SortAssetNodes(nodes)

	assert.Equal(t, []string{"SS-001", "SS-002", "SS-010", "SS-012", "SS-T01"}, assetIDs(nodes))
}

func TestSortAssetNodes_IsTotalForShuffledInput(t *testing.T) {
	want := []string{"A", "A-1", "A-01", "A-2", "A-10", "A-B", "B", "b"}
	random := rand.New(rand.NewSource(1))

	for range 20 {
		nodes := []*AssetNode{}
		for _, id := range want {
			nodes = append(nodes, node(id, valueobject.AssetTypeTransformer))
		}
		random.Shuffle(len(nodes), func(i, j int) { nodes[i], nodes[j] = nodes[j], nodes[i] })

		SortAssetNodes(nodes)

		assert.Equal(t, want, assetIDs(nodes))
	}
}

func TestSortAssetNodes_EmptyAndNil(t *testing.T) {
	assert.NotPanics(t, func() { SortAssetNodes(nil) })
	empty := []*AssetNode{}
	SortAssetNodes(empty)
	assert.Empty(t, empty)
}
