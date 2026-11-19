package assetimport

import (
	"testing"

	"smart-grid-asset-registry/domain/errorcode"
	"smart-grid-asset-registry/domain/valueobject"

	"github.com/stretchr/testify/assert"
)

// line builds a valid data line with the given ID, parent and type. The
// serial number is left blank so no serial warning interferes.
func line(assetID, parentAssetID, assetType string) string {
	return assetID + "," + parentAssetID + "," + assetType + ",Name " + assetID + ",IN_SERVICE,22,,,,,"
}

func TestHierarchy_ParentPresence(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		id    string
		want  []Issue
	}{
		{
			"substation cannot have a parent",
			[]string{line("SS-1", "", "SUBSTATION"), line("SS-HP", "SS-1", "SUBSTATION")},
			"SS-HP",
			[]Issue{{Code: errorcode.SubstationHasParent, Args: []any{"SS-1"}}},
		},
		{
			"non-substation needs a parent",
			[]string{line("TX-NP", "", "TRANSFORMER")},
			"TX-NP",
			[]Issue{{Code: errorcode.ParentRequired, Args: []any{"TRANSFORMER"}}},
		},
		{
			"asset cannot be its own parent",
			[]string{line("SS-1", "", "SUBSTATION"), line("SWB-1", "SWB-1", "SWITCHBOARD")},
			"SWB-1",
			[]Issue{{Code: errorcode.SelfParent}},
		},
		{
			"parent neither stored nor in file",
			[]string{line("PNL-M", "SWB-404", "SWITCHBOARD_PANEL")},
			"PNL-M",
			[]Issue{{Code: errorcode.ParentNotFound, Args: []any{"SWB-404"}}},
		},
		{
			"a self-looking parent ID that names a missing asset is PARENT_NOT_FOUND",
			[]string{line("SWB-S", "SWB-SELF", "SWITCHBOARD")},
			"SWB-S",
			[]Issue{{Code: errorcode.ParentNotFound, Args: []any{"SWB-SELF"}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := validateLines(t, emptyStore(), tt.lines...)

			assert.Equal(t, tt.want, issuesOf(t, report, tt.id))
		})
	}
}

func TestHierarchy_StoredParentIsAccepted(t *testing.T) {
	store := storeWith(storedAsset("SS-STORED", "", valueobject.AssetTypeSubstation))

	report := validateLines(t, store, line("TX-1", "SS-STORED", "TRANSFORMER"))

	assert.Empty(t, report.Rejected)
	assert.Equal(t, []string{"TX-1"}, acceptedIDs(report))
}

func TestHierarchy_InvalidParentType(t *testing.T) {
	t.Run("in-file parent", func(t *testing.T) {
		report := validateLines(t, emptyStore(),
			line("SS-1", "", "SUBSTATION"),
			line("PNL-BP", "SS-1", "SWITCHBOARD_PANEL"),
		)

		assert.Equal(t, []Issue{{
			Code: errorcode.InvalidParentType,
			Args: []any{"SWITCHBOARD_PANEL", "SWITCHBOARD", "SS-1", "SUBSTATION"},
		}}, issuesOf(t, report, "PNL-BP"))
	})

	t.Run("stored parent", func(t *testing.T) {
		store := storeWith(
			storedAsset("SS-1", "", valueobject.AssetTypeSubstation),
			storedAsset("SWB-1", "SS-1", valueobject.AssetTypeSwitchboard),
		)

		report := validateLines(t, store, line("TX-BP", "SWB-1", "TRANSFORMER"))

		assert.Equal(t, []Issue{{
			Code: errorcode.InvalidParentType,
			Args: []any{"TRANSFORMER", "SUBSTATION", "SWB-1", "SWITCHBOARD"},
		}}, issuesOf(t, report, "TX-BP"))
	})

	t.Run("parent with an unsupported type is only PARENT_REJECTED", func(t *testing.T) {
		report := validateLines(t, emptyStore(),
			line("GEN-1", "", "GENERATOR"),
			line("TX-1", "GEN-1", "TRANSFORMER"),
		)

		assert.Equal(t, []string{errorcode.InvalidAssetType}, codesOf(report)["GEN-1"])
		assert.Equal(t, []Issue{{Code: errorcode.ParentRejected, Args: []any{"GEN-1"}}}, issuesOf(t, report, "TX-1"))
	})
}

func TestHierarchy_Cycles(t *testing.T) {
	t.Run("two-cycle: every member, with every applicable error", func(t *testing.T) {
		report := validateLines(t, emptyStore(),
			line("SWB-CY-A", "SWB-CY-B", "SWITCHBOARD"),
			line("SWB-CY-B", "SWB-CY-A", "SWITCHBOARD"),
		)

		assert.Equal(t, []Issue{
			{Code: errorcode.InvalidParentType, Args: []any{"SWITCHBOARD", "SUBSTATION", "SWB-CY-B", "SWITCHBOARD"}},
			{Code: errorcode.CycleDetected, Args: []any{"SWB-CY-A", "SWB-CY-A, SWB-CY-B"}},
		}, issuesOf(t, report, "SWB-CY-A"))
		assert.Equal(t, []string{errorcode.InvalidParentType, errorcode.CycleDetected}, codesOf(report)["SWB-CY-B"])
	})

	t.Run("three-cycle: every member", func(t *testing.T) {
		report := validateLines(t, emptyStore(),
			line("SWB-CY-C", "PNL-CY-C", "SWITCHBOARD"),
			line("PNL-CY-C", "PNL-CY-D", "SWITCHBOARD_PANEL"),
			line("PNL-CY-D", "SWB-CY-C", "SWITCHBOARD_PANEL"),
		)

		codes := codesOf(report)
		assert.Equal(t, []string{errorcode.InvalidParentType, errorcode.CycleDetected}, codes["SWB-CY-C"])
		assert.Equal(t, []string{errorcode.InvalidParentType, errorcode.CycleDetected}, codes["PNL-CY-C"])
		assert.Equal(t, []Issue{
			{Code: errorcode.CycleDetected, Args: []any{"PNL-CY-D", "PNL-CY-C, PNL-CY-D, SWB-CY-C"}},
		}, issuesOf(t, report, "PNL-CY-D"), "a member whose own parent type is right still gets CYCLE_DETECTED")
	})

	t.Run("rows below a cycle are PARENT_REJECTED; members never are", func(t *testing.T) {
		report := validateLines(t, emptyStore(),
			line("SWB-CY-A", "SWB-CY-B", "SWITCHBOARD"),
			line("SWB-CY-B", "SWB-CY-A", "SWITCHBOARD"),
			line("PNL-BELOW", "SWB-CY-A", "SWITCHBOARD_PANEL"),
		)

		codes := codesOf(report)
		assert.Equal(t, []Issue{{Code: errorcode.ParentRejected, Args: []any{"SWB-CY-A"}}}, issuesOf(t, report, "PNL-BELOW"))
		assert.NotContains(t, codes["SWB-CY-A"], errorcode.ParentRejected)
		assert.NotContains(t, codes["SWB-CY-B"], errorcode.ParentRejected)
	})
}

func TestHierarchy_ParentRejectedPropagates(t *testing.T) {
	t.Run("down a three-level chain", func(t *testing.T) {
		report := validateLines(t, emptyStore(),
			line("PNL-1", "SWB-1", "SWITCHBOARD_PANEL"),
			line("SWB-1", "SS-HP", "SWITCHBOARD"),
			line("SS-HP", "SS-ROOT", "SUBSTATION"),
			line("SS-ROOT", "", "SUBSTATION"),
		)

		assert.Equal(t, []Issue{{Code: errorcode.ParentRejected, Args: []any{"SS-HP"}}}, issuesOf(t, report, "SWB-1"))
		assert.Equal(t, []Issue{{Code: errorcode.ParentRejected, Args: []any{"SWB-1"}}}, issuesOf(t, report, "PNL-1"))
		assert.Equal(t, []string{"SS-ROOT"}, acceptedIDs(report))
	})

	t.Run("children of a duplicated parent ID", func(t *testing.T) {
		report := validateLines(t, emptyStore(),
			line("SS-D", "", "SUBSTATION"),
			line("SS-D", "", "SUBSTATION"),
			line("TX-1", "SS-D", "TRANSFORMER"),
		)

		assert.Equal(t, []Issue{{Code: errorcode.ParentRejected, Args: []any{"SS-D"}}}, issuesOf(t, report, "TX-1"))
	})

	t.Run("an in-file parent with field errors", func(t *testing.T) {
		report := validateLines(t, emptyStore(),
			"SS-1,,SUBSTATION,,IN_SERVICE,22,,,,,",
			line("TX-1", "SS-1", "TRANSFORMER"),
		)

		assert.Equal(t, []string{errorcode.AssetNameRequired}, codesOf(report)["SS-1"])
		assert.Equal(t, []Issue{{Code: errorcode.ParentRejected, Args: []any{"SS-1"}}}, issuesOf(t, report, "TX-1"))
	})
}

func TestHierarchy_SkippedForRowsWithFieldErrors(t *testing.T) {
	report := validateLines(t, emptyStore(), "TX-1,,TRANSFORMER,,IN_SERVICE,22,,,,,")

	assert.Equal(t, []string{errorcode.AssetNameRequired}, codesOf(report)["TX-1"], "no PARENT_REQUIRED")
}

func TestHierarchy_ChildBeforeParentAndGrandparentIsAccepted(t *testing.T) {
	report := validateLines(t, emptyStore(),
		line("PNL-1", "SWB-1", "SWITCHBOARD_PANEL"),
		line("SWB-1", "SS-1", "SWITCHBOARD"),
		line("SS-1", "", "SUBSTATION"),
	)

	assert.Empty(t, report.Rejected)
	assert.ElementsMatch(t, []string{"PNL-1", "SWB-1", "SS-1"}, acceptedIDs(report))
}
