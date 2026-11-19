package assetimport

import (
	"testing"

	"smart-grid-asset-registry/domain/errorcode"
	"smart-grid-asset-registry/domain/valueobject"

	"github.com/stretchr/testify/assert"
)

// withSerial builds a valid data line with a serial number.
func withSerial(assetID, parentAssetID, assetType, serial string) string {
	return assetID + "," + parentAssetID + "," + assetType + ",Name " + assetID + ",IN_SERVICE,22,,,," + serial + ","
}

func TestSerialWarnings_TwoAcceptedRowsWarnEachOther(t *testing.T) {
	report := validateLines(t, emptyStore(),
		withSerial("SS-1", "", "SUBSTATION", "S-SS"),
		withSerial("TX-D1", "SS-1", "TRANSFORMER", "SER-TX-DUP"),
		withSerial("TX-D2", "SS-1", "TRANSFORMER", "SER-TX-DUP"),
	)

	assert.Equal(t, []Warning{
		{Row: 3, AssetID: "TX-D1", Issue: Issue{Code: errorcode.DuplicateSerialNumber, Args: []any{"SER-TX-DUP", "TX-D2"}}},
		{Row: 4, AssetID: "TX-D2", Issue: Issue{Code: errorcode.DuplicateSerialNumber, Args: []any{"SER-TX-DUP", "TX-D1"}}},
	}, report.Warnings)
	assert.Equal(t, 2, report.Counts.WarningRows)
	assert.Len(t, report.Accepted, 3, "warnings never change acceptance")
}

func TestSerialWarnings_StoredAssetWithAnotherID(t *testing.T) {
	stored := storedAsset("TX-OLD", "SS-1", valueobject.AssetTypeTransformer)
	serial := "SER-1"
	stored.SerialNumber = &serial
	store := storeWith(storedAsset("SS-1", "", valueobject.AssetTypeSubstation), stored)

	report := validateLines(t, store, withSerial("TX-NEW", "SS-1", "TRANSFORMER", "SER-1"))

	assert.Equal(t, []Warning{
		{Row: 2, AssetID: "TX-NEW", Issue: Issue{Code: errorcode.DuplicateSerialNumber, Args: []any{"SER-1", "TX-OLD"}}},
	}, report.Warnings)
}

func TestSerialWarnings_OtherIDsAreSortedAndUnique(t *testing.T) {
	stored := storedAsset("TX-A", "SS-1", valueobject.AssetTypeTransformer)
	serial := "SER-1"
	stored.SerialNumber = &serial
	store := storeWith(storedAsset("SS-1", "", valueobject.AssetTypeSubstation), stored)

	report := validateLines(t, store,
		withSerial("TX-C", "SS-1", "TRANSFORMER", "SER-1"),
		withSerial("TX-B", "SS-1", "TRANSFORMER", "SER-1"),
	)

	assert.Equal(t, []any{"SER-1", "TX-A, TX-B"}, report.Warnings[0].Issue.Args)
	assert.Equal(t, []any{"SER-1", "TX-A, TX-C"}, report.Warnings[1].Issue.Args)
}

func TestSerialWarnings_RejectedRowsNeitherWarnNorTriggerWarnings(t *testing.T) {
	report := validateLines(t, emptyStore(),
		withSerial("SS-1", "", "SUBSTATION", "SER-1"),
		withSerial("TX-BAD", "", "TRANSFORMER", "SER-1"),
	)

	assert.Contains(t, codesOf(report), "TX-BAD")
	assert.Empty(t, report.Warnings)
	assert.Zero(t, report.Counts.WarningRows)
}

func TestSerialWarnings_BlankSerialsAreIgnored(t *testing.T) {
	report := validateLines(t, emptyStore(),
		line("SS-1", "", "SUBSTATION"),
		line("SS-2", "", "SUBSTATION"),
	)

	assert.Empty(t, report.Warnings)
}
