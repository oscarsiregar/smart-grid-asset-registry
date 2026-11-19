package asset_import

import (
	"encoding/json"
	"testing"
	"time"

	"smart-grid-asset-registry/domain/assetimport"
	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/errorcode"
	"smart-grid-asset-registry/domain/valueobject"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleReport() assetimport.Report {
	assetID := "SWB-CY-A"
	return assetimport.Report{
		Counts: assetimport.Counts{Total: 4, ToImport: 2, Unchanged: 0, Rejected: 2, WarningRows: 1},
		Rejected: []assetimport.RejectedRow{
			{Row: 93, AssetID: &assetID, Errors: []assetimport.Issue{
				{Code: errorcode.InvalidParentType, Args: []any{"SWITCHBOARD", "SUBSTATION", "SWB-CY-B", "SWITCHBOARD"}},
				{Code: errorcode.CycleDetected, Args: []any{"SWB-CY-A", "SWB-CY-A, SWB-CY-B"}},
			}},
			{Row: 182, Errors: []assetimport.Issue{{Code: errorcode.AssetIDRequired}}},
		},
		Warnings: []assetimport.Warning{
			{Row: 89, AssetID: "TX-D1", Issue: assetimport.Issue{Code: errorcode.DuplicateSerialNumber, Args: []any{"SER-TX-DUP", "TX-D2"}}},
		},
	}
}

func TestToImportResultResponse_ValidatedOnly(t *testing.T) {
	got := toImportResultResponse(valueobject.ImportOutcomeValidatedOnly, valueobject.ImportModeStrict, "grid.csv", sampleReport(), nil)

	raw, err := json.Marshal(got)
	require.NoError(t, err)

	assert.JSONEq(t, `{
		"outcome": "VALIDATED_ONLY",
		"isCommitted": false,
		"mode": "STRICT",
		"fileName": "grid.csv",
		"counts": {"total": 4, "imported": 2, "unchanged": 0, "rejected": 2, "warningRows": 1},
		"rejectedRows": [
			{"row": 93, "assetId": "SWB-CY-A", "errors": [
				{"code": "INVALID_PARENT_TYPE", "message": "A SWITCHBOARD must have a SUBSTATION as parent, but SWB-CY-B is a SWITCHBOARD."},
				{"code": "CYCLE_DETECTED", "message": "SWB-CY-A is part of a parent cycle: SWB-CY-A, SWB-CY-B."}
			]},
			{"row": 182, "assetId": null, "errors": [{"code": "ASSET_ID_REQUIRED", "message": "asset_id is required."}]}
		],
		"warnings": [
			{"row": 89, "assetId": "TX-D1", "code": "DUPLICATE_SERIAL_NUMBER", "message": "Serial number SER-TX-DUP is also used by TX-D2."}
		],
		"import": null
	}`, string(raw))
}

func TestToImportResultResponse_EmptyListsAreNotNull(t *testing.T) {
	got := toImportResultResponse(valueobject.ImportOutcomeValidatedOnly, valueobject.ImportModePartial, "a.csv", assetimport.Report{}, nil)

	raw, err := json.Marshal(got)
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"rejectedRows":[]`)
	assert.Contains(t, string(raw), `"warnings":[]`)
}

func TestToImportResultResponse_CommittedWithRecord(t *testing.T) {
	record := &entity.AssetImport{
		ID:        uuid.MustParse("6f1c2c1e-6a4e-4b8e-9a52-3d1f2b9c0a11"),
		CreatedAt: time.Date(2026, 11, 19, 17, 15, 0, 0, time.FixedZone("WIB", 7*60*60)),
	}

	got := toImportResultResponse(valueobject.ImportOutcomeCommitted, valueobject.ImportModePartial, "a.csv", assetimport.Report{}, record)

	assert.True(t, got.IsCommitted)
	require.NotNil(t, got.Import)
	assert.Equal(t, "6f1c2c1e-6a4e-4b8e-9a52-3d1f2b9c0a11", got.Import.ID)
	assert.Equal(t, "2026-11-19T10:15:00Z", got.Import.CreatedAt, "RFC 3339 in UTC")
}

func TestToImportResultResponse_OnlyCommittedIsCommitted(t *testing.T) {
	for _, outcome := range []valueobject.ImportOutcome{
		valueobject.ImportOutcomeValidatedOnly,
		valueobject.ImportOutcomeNothingToCommit,
		valueobject.ImportOutcomeStrictBlocked,
	} {
		got := toImportResultResponse(outcome, valueobject.ImportModePartial, "a.csv", assetimport.Report{}, nil)
		assert.Falsef(t, got.IsCommitted, "outcome %s", outcome)
	}
}
