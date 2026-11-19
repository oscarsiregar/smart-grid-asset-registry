package assetimport

import (
	"testing"

	"smart-grid-asset-registry/domain/errorcode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	reSubstation  = "SS-1,,SUBSTATION,Sub 1,IN_SERVICE,22,,,,SUB-1,15/2/09"
	reTransformer = "TX-1,SS-1,TRANSFORMER,Tx 1,IN_SERVICE,22,1000,Arcwell,TXR,TX-SER-1,12/3/10"
)

// storedFrom stores the assets the given lines produce against an empty store,
// exactly as a committed import would have saved them.
func storedFrom(t *testing.T, lines ...string) StoredAssets {
	t.Helper()
	report := validateLines(t, emptyStore(), lines...)
	require.Empty(t, report.Rejected, "store fixture rows must be valid")
	return storeWith(report.Accepted...)
}

func TestReimport_IdenticalRowIsUnchanged(t *testing.T) {
	stored := storedFrom(t, reSubstation, reTransformer)

	// written differently but equal by value: 22.0 vs 22, ISO date, padding
	report := validateLines(t, stored,
		"SS-1,,SUBSTATION, Sub 1 ,IN_SERVICE,22.0,,,,SUB-1,2009-02-15",
		reTransformer,
	)

	assert.Empty(t, report.Rejected)
	assert.Empty(t, report.Accepted, "unchanged rows are not imported")
	assert.Equal(t, []int{2, 3}, report.UnchangedRows)
	assert.Equal(t, Counts{Total: 2, ToImport: 0, Unchanged: 2, Rejected: 0}, report.Counts)
}

func TestReimport_ChangedStatusIsConflict(t *testing.T) {
	stored := storedFrom(t, reSubstation, reTransformer)

	report := validateLines(t, stored,
		"TX-1,SS-1,TRANSFORMER,Tx 1,MAINTENANCE,22,1000,Arcwell,TXR,TX-SER-1,12/3/10",
	)

	require.Len(t, report.Rejected, 1)
	assert.Equal(t, []Issue{{Code: errorcode.AssetIDConflict, Args: []any{"TX-1", "operational_status"}}},
		report.Rejected[0].Errors)
	assert.Equal(t, 1, report.Counts.Rejected)
	assert.Zero(t, report.Counts.Unchanged)
}

func TestReimport_ConflictListsEveryDifferingField(t *testing.T) {
	stored := storedFrom(t, reSubstation, reTransformer)

	report := validateLines(t, stored,
		"TX-1,SS-1,TRANSFORMER,Tx One,IN_SERVICE,22,2000,Arcwell,TXR,TX-SER-1,",
	)

	assert.Equal(t, []Issue{{Code: errorcode.AssetIDConflict, Args: []any{"TX-1", "asset_name, rating_kva, commissioned_date"}}},
		issuesOf(t, report, "TX-1"))
}

func TestReimport_ChildOfConflictingRowIsParentRejected(t *testing.T) {
	stored := storedFrom(t, reSubstation)

	report := validateLines(t, stored,
		"SS-1,,SUBSTATION,Renamed,IN_SERVICE,22,,,,SUB-1,15/2/09",
		"TX-2,SS-1,TRANSFORMER,Tx 2,IN_SERVICE,22,1000,,,,",
	)

	assert.Equal(t, map[string][]string{
		"SS-1": {errorcode.AssetIDConflict},
		"TX-2": {errorcode.ParentRejected},
	}, codesOf(report))
}

func TestReimport_UnchangedRowIsAValidParent(t *testing.T) {
	stored := storedFrom(t, reSubstation)

	report := validateLines(t, stored,
		"TX-2,SS-1,TRANSFORMER,Tx 2,IN_SERVICE,22,1000,,,,",
		reSubstation,
	)

	assert.Empty(t, report.Rejected)
	require.Len(t, report.Accepted, 1)
	assert.Equal(t, "TX-2", report.Accepted[0].AssetID)
	assert.Equal(t, []int{3}, report.UnchangedRows)
}

func TestReimport_RowWithInvalidFieldsIsNotComparedToTheStore(t *testing.T) {
	stored := storedFrom(t, reSubstation)

	report := validateLines(t, stored, "SS-1,,SUBSTATION,Sub 1,IN_SERVICE,abc,,,,SUB-1,15/2/09")

	assert.Equal(t, map[string][]string{"SS-1": {errorcode.InvalidNumber}}, codesOf(report))
}

// TestReimport_SampleAgainstItsOwnImport: importing the sample a second time
// adds nothing (US4 independent test).
func TestReimport_SampleAgainstItsOwnImport(t *testing.T) {
	first := Validate(parseSample(t), emptyStore())

	second := Validate(parseSample(t), storeWith(first.Accepted...))

	assert.Empty(t, second.Accepted)
	assert.Equal(t, Counts{Total: 222, ToImport: 0, Unchanged: 205, Rejected: 17, WarningRows: 2}, second.Counts)
	assert.Equal(t, codesOfAll(first), codesOfAll(second), "the same 17 rows are rejected for the same reasons")
}

func codesOfAll(report Report) map[int][]string {
	codes := map[int][]string{}
	for _, rejected := range report.Rejected {
		for _, issue := range rejected.Errors {
			codes[rejected.Row] = append(codes[rejected.Row], issue.Code)
		}
	}
	return codes
}
