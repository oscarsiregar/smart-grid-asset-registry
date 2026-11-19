package entity

import (
	"math/rand"
	"testing"

	"smart-grid-asset-registry/domain/valueobject"

	"github.com/stretchr/testify/assert"
)

func namedNode(assetID, assetName string, assetType valueobject.AssetType) *AssetNode {
	return &AssetNode{AssetID: assetID, AssetName: assetName, AssetType: assetType}
}

func TestCheckIsExactMatch(t *testing.T) {
	panel := namedNode("PNL-001-1-01", "Aurora Panel A1", valueobject.AssetTypeSwitchboardPanel)

	tests := []struct {
		name string
		text string
		want bool
	}{
		{"the whole ID", "PNL-001-1-01", true},
		{"the whole ID, other letter case", "pnl-001-1-01", true},
		{"the whole name", "Aurora Panel A1", true},
		{"the whole name, other letter case", "AURORA PANEL a1", true},
		{"part of the ID", "PNL-001", false},
		{"part of the name", "Aurora", false},
		{"unrelated text", "SS-001", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, CheckIsExactMatch(panel, tt.text))
		})
	}
}

func TestSortSearchMatches_ExactMatchesFirstThenExplorerOrder(t *testing.T) {
	nodes := []*AssetNode{
		namedNode("PNL-001-1-10", "Panel 10", valueobject.AssetTypeSwitchboardPanel),
		namedNode("SWB-001-1", "Aurora 22kV Switchboard A", valueobject.AssetTypeSwitchboard),
		namedNode("TX-001-1", "Aurora Transformer T1", valueobject.AssetTypeTransformer),
		namedNode("SS-001", "Aurora", valueobject.AssetTypeSubstation),
		namedNode("PNL-001-1-2", "Panel 2", valueobject.AssetTypeSwitchboardPanel),
		namedNode("TX-002-1", "aurora", valueobject.AssetTypeTransformer),
	}

	SortSearchMatches(nodes, "Aurora")

	assert.Equal(t, []string{
		// exact (name equals "Aurora", letter case ignored), in explorer order
		"SS-001", "TX-002-1",
		// the rest, in explorer order: type rank, then natural ID
		"TX-001-1", "SWB-001-1", "PNL-001-1-2", "PNL-001-1-10",
	}, assetIDs(nodes))
}

func TestSortSearchMatches_FullIDIsFirst(t *testing.T) {
	nodes := []*AssetNode{
		namedNode("SS-001", "Aurora", valueobject.AssetTypeSubstation),
		namedNode("PNL-001-1-01", "Aurora Panel A1", valueobject.AssetTypeSwitchboardPanel),
		namedNode("TX-001-1", "Aurora Transformer T1", valueobject.AssetTypeTransformer),
	}

	SortSearchMatches(nodes, "pnl-001-1-01")

	assert.Equal(t, "PNL-001-1-01", nodes[0].AssetID)
}

func TestSortSearchMatches_IsTotalForShuffledInput(t *testing.T) {
	want := []string{"SS-1", "SS-2", "TX-1", "TX-10", "PNL-1", "PNL-01", "PNL-2"}
	random := rand.New(rand.NewSource(3))

	for range 20 {
		nodes := []*AssetNode{
			namedNode("PNL-2", "x", valueobject.AssetTypeSwitchboardPanel),
			namedNode("TX-10", "x", valueobject.AssetTypeTransformer),
			namedNode("SS-2", "x", valueobject.AssetTypeSubstation),
			namedNode("PNL-01", "x", valueobject.AssetTypeSwitchboardPanel),
			namedNode("TX-1", "x", valueobject.AssetTypeTransformer),
			namedNode("SS-1", "x", valueobject.AssetTypeSubstation),
			namedNode("PNL-1", "x", valueobject.AssetTypeSwitchboardPanel),
		}
		random.Shuffle(len(nodes), func(i, j int) { nodes[i], nodes[j] = nodes[j], nodes[i] })

		SortSearchMatches(nodes, "no exact match")

		assert.Equal(t, want, assetIDs(nodes))
	}
}

func TestSortSearchMatches_EmptyAndNil(t *testing.T) {
	assert.NotPanics(t, func() { SortSearchMatches(nil, "x") })
	empty := []*AssetNode{}
	SortSearchMatches(empty, "x")
	assert.Empty(t, empty)
}
