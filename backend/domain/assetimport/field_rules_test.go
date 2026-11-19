package assetimport

import (
	"testing"
	"time"

	"smart-grid-asset-registry/domain/errorcode"
	"smart-grid-asset-registry/domain/valueobject"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const headerFieldCount = 11

// validValues is a complete, valid transformer row.
func validValues() map[Column]string {
	return map[Column]string{
		ColumnAssetID:           "TX-1",
		ColumnParentAssetID:     "SS-1",
		ColumnAssetType:         "TRANSFORMER",
		ColumnAssetName:         "Transformer 1",
		ColumnOperationalStatus: "IN_SERVICE",
		ColumnVoltageKV:         "22",
		ColumnRatingKVA:         "2000",
		ColumnManufacturer:      "Arcwell",
		ColumnModel:             "TXR-2000",
		ColumnSerialNumber:      "TX0001",
		ColumnCommissionedDate:  "12/3/10",
	}
}

func candidateWith(overrides map[Column]string) *candidate {
	values := validValues()
	for column, value := range overrides {
		values[column] = value
	}
	return newCandidate(Row{Number: 5, FieldCount: headerFieldCount, Values: values})
}

func TestCheckFields(t *testing.T) {
	tests := []struct {
		name      string
		overrides map[Column]string
		want      []Issue
	}{
		{"valid row", nil, nil},
		{"asset_id required", map[Column]string{ColumnAssetID: " "}, []Issue{{Code: errorcode.AssetIDRequired}}},
		{"asset_type required", map[Column]string{ColumnAssetType: ""}, []Issue{{Code: errorcode.AssetTypeRequired}}},
		{"asset_name required", map[Column]string{ColumnAssetName: ""}, []Issue{{Code: errorcode.AssetNameRequired}}},
		{"unknown asset type", map[Column]string{ColumnAssetType: "GENERATOR"},
			[]Issue{{Code: errorcode.InvalidAssetType, Args: []any{"GENERATOR"}}}},
		{"asset type is case-sensitive", map[Column]string{ColumnAssetType: "transformer"},
			[]Issue{{Code: errorcode.InvalidAssetType, Args: []any{"transformer"}}}},
		{"blank status is accepted", map[Column]string{ColumnOperationalStatus: ""}, nil},
		{"unknown status", map[Column]string{ColumnOperationalStatus: "BROKEN"},
			[]Issue{{Code: errorcode.InvalidOperationalStatus, Args: []any{"BROKEN"}}}},
		{"voltage not a number", map[Column]string{ColumnVoltageKV: "22kV"},
			[]Issue{{Code: errorcode.InvalidNumber, Args: []any{"voltage_kv", "22kV"}}}},
		{"rating not a number", map[Column]string{ColumnRatingKVA: "1e3"},
			[]Issue{{Code: errorcode.InvalidNumber, Args: []any{"rating_kva", "1e3"}}}},
		{"decimal numbers", map[Column]string{ColumnVoltageKV: "0.4", ColumnRatingKVA: ".5"}, nil},
		{"negative rating", map[Column]string{ColumnRatingKVA: "-500"},
			[]Issue{{Code: errorcode.NegativeRating, Args: []any{"-500"}}}},
		{"negative voltage is accepted (clarification Q2)", map[Column]string{ColumnVoltageKV: "-22"}, nil},
		{"future date is accepted (clarification Q2)", map[Column]string{ColumnCommissionedDate: "2099-01-01"}, nil},
		{"iso date", map[Column]string{ColumnCommissionedDate: "2009-02-15"}, nil},
		{"impossible day", map[Column]string{ColumnCommissionedDate: "31/2/20"},
			[]Issue{{Code: errorcode.InvalidDate, Args: []any{"31/2/20"}}}},
		{"impossible iso date", map[Column]string{ColumnCommissionedDate: "2026-13-41"},
			[]Issue{{Code: errorcode.InvalidDate, Args: []any{"2026-13-41"}}}},
		{"unsupported date format", map[Column]string{ColumnCommissionedDate: "Feb 15 2009"},
			[]Issue{{Code: errorcode.InvalidDate, Args: []any{"Feb 15 2009"}}}},
		{"several problems are all listed", map[Column]string{ColumnAssetName: "", ColumnOperationalStatus: "BROKEN", ColumnRatingKVA: "-1"},
			[]Issue{
				{Code: errorcode.AssetNameRequired},
				{Code: errorcode.InvalidOperationalStatus, Args: []any{"BROKEN"}},
				{Code: errorcode.NegativeRating, Args: []any{"-1"}},
			}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := candidateWith(tt.overrides)

			checkFields(c, headerFieldCount)

			assert.Equal(t, tt.want, c.fieldIssues)
			assert.Equal(t, tt.want == nil, c.asset != nil, "an asset is built exactly when the fields are valid")
		})
	}
}

func TestCheckFields_MalformedRowReportsOnlyTheFieldCount(t *testing.T) {
	c := newCandidate(Row{Number: 9, FieldCount: 3, Values: map[Column]string{
		ColumnAssetID: "SS-1", ColumnParentAssetID: "", ColumnAssetType: "SUBSTATION",
	}})

	checkFields(c, headerFieldCount)

	// Values of a row with a missing or extra field are misaligned, so their
	// field rules would only produce noise.
	assert.Equal(t, []Issue{{Code: errorcode.MalformedRow, Args: []any{3, headerFieldCount}}}, c.fieldIssues)
	assert.Nil(t, c.asset)
}

func TestCheckFields_BuildsTheAsset(t *testing.T) {
	c := candidateWith(map[Column]string{ColumnCommissionedDate: "15/2/09", ColumnManufacturer: ""})

	checkFields(c, headerFieldCount)

	require.NotNil(t, c.asset)
	asset := c.asset
	assert.Equal(t, "TX-1", asset.AssetID)
	require.NotNil(t, asset.ParentAssetID)
	assert.Equal(t, "SS-1", *asset.ParentAssetID)
	assert.Equal(t, valueobject.AssetTypeTransformer, asset.AssetType)
	assert.Equal(t, "Transformer 1", asset.AssetName)
	require.NotNil(t, asset.OperationalStatus)
	assert.Equal(t, valueobject.OperationalStatusInService, *asset.OperationalStatus)
	assert.Equal(t, 22.0, *asset.VoltageKV)
	assert.Equal(t, 2000.0, *asset.RatingKVA)
	assert.Nil(t, asset.Manufacturer, "blank becomes nil")
	assert.Equal(t, "TXR-2000", *asset.Model)
	assert.Equal(t, "TX0001", *asset.SerialNumber)
	assert.Equal(t, time.Date(2009, time.February, 15, 0, 0, 0, 0, time.UTC), *asset.CommissionedDate,
		"two-digit years mean 20yy, day first")
}

func TestCheckFields_BlankOptionalValuesBecomeNil(t *testing.T) {
	c := newCandidate(Row{Number: 2, FieldCount: headerFieldCount, Values: map[Column]string{
		ColumnAssetID: "SS-1", ColumnAssetType: "SUBSTATION", ColumnAssetName: "Sub",
	}})

	checkFields(c, headerFieldCount)

	require.NotNil(t, c.asset)
	assert.Nil(t, c.asset.ParentAssetID)
	assert.Nil(t, c.asset.OperationalStatus)
	assert.Nil(t, c.asset.VoltageKV)
	assert.Nil(t, c.asset.RatingKVA)
	assert.Nil(t, c.asset.SerialNumber)
	assert.Nil(t, c.asset.CommissionedDate)
}

func TestCheckDuplicateAssetIDs_RejectsEveryRowSharingTheID(t *testing.T) {
	rows := []*candidate{
		newCandidate(Row{Number: 2, FieldCount: headerFieldCount, Values: map[Column]string{ColumnAssetID: "TX-1"}}),
		newCandidate(Row{Number: 3, FieldCount: headerFieldCount, Values: map[Column]string{ColumnAssetID: "TX-2"}}),
		newCandidate(Row{Number: 9, FieldCount: headerFieldCount, Values: map[Column]string{ColumnAssetID: "TX-1"}}),
		newCandidate(Row{Number: 4, FieldCount: headerFieldCount, Values: map[Column]string{ColumnAssetID: ""}}),
		newCandidate(Row{Number: 5, FieldCount: headerFieldCount, Values: map[Column]string{ColumnAssetID: ""}}),
	}

	checkDuplicateAssetIDs(rows)

	want := []Issue{{Code: errorcode.DuplicateAssetID, Args: []any{"TX-1", "2, 9"}}}
	assert.Equal(t, want, rows[0].fieldIssues)
	assert.Equal(t, want, rows[2].fieldIssues)
	assert.Empty(t, rows[1].fieldIssues)
	assert.Empty(t, rows[3].fieldIssues, "blank IDs are not duplicates of each other")
	assert.Empty(t, rows[4].fieldIssues)
}
