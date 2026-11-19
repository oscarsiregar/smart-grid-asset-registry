package assetimport

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCandidate_TrimsEveryField(t *testing.T) {
	row := Row{Number: 7, FieldCount: 11, Values: map[Column]string{
		ColumnAssetID:           "  SS-1 ",
		ColumnParentAssetID:     "\t",
		ColumnAssetType:         " SUBSTATION",
		ColumnAssetName:         "  Dual Node  B  ",
		ColumnOperationalStatus: "IN_SERVICE ",
		ColumnVoltageKV:         " 22 ",
		ColumnRatingKVA:         "   ",
		ColumnManufacturer:      " Arcwell ",
		ColumnModel:             " TXR ",
		ColumnSerialNumber:      " S-1 ",
		ColumnCommissionedDate:  " 15/2/09 ",
	}}

	c := newCandidate(row)

	assert.Equal(t, row, c.row)
	assert.Equal(t, "SS-1", c.assetID)
	assert.Equal(t, "", c.parentAssetID, "whitespace-only counts as empty")
	assert.Equal(t, "SUBSTATION", c.assetType)
	assert.Equal(t, "Dual Node  B", c.assetName, "inner whitespace is kept")
	assert.Equal(t, "IN_SERVICE", c.operationalStatus)
	assert.Equal(t, "22", c.voltageKV)
	assert.Equal(t, "", c.ratingKVA)
	assert.Equal(t, "Arcwell", c.manufacturer)
	assert.Equal(t, "TXR", c.model)
	assert.Equal(t, "S-1", c.serialNumber)
	assert.Equal(t, "15/2/09", c.commissionedDate)
	assert.False(t, c.rejected())
}

func TestNewCandidate_MissingValuesAreEmpty(t *testing.T) {
	c := newCandidate(Row{Number: 2, FieldCount: 1, Values: map[Column]string{ColumnAssetID: "X"}})

	assert.Equal(t, "X", c.assetID)
	assert.Equal(t, "", c.assetName)
	assert.Equal(t, "", c.commissionedDate)
}

func TestCandidate_Rejected(t *testing.T) {
	fieldError := &candidate{fieldIssues: []Issue{{Code: "X"}}}
	hierarchyError := &candidate{hierarchyIssues: []Issue{{Code: "Y"}}}

	assert.True(t, fieldError.rejected())
	assert.True(t, hierarchyError.rejected())
}
