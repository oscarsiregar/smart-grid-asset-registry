package assetimport

import (
	"strings"
	"testing"

	"smart-grid-asset-registry/domain/errorcode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const substationLine = "SS-1,,SUBSTATION,Sub One,IN_SERVICE,22,,,,S-1,15/2/09"

func TestParse_ValidFile(t *testing.T) {
	file, fileErr := Parse(strings.NewReader(csvText(substationLine)), testMaxBytes)

	require.Nil(t, fileErr)
	assert.Equal(t, 11, file.HeaderFieldCount)
	require.Len(t, file.Rows, 1)
	assert.Equal(t, 2, file.Rows[0].Number, "header is line 1")
	assert.Equal(t, 11, file.Rows[0].FieldCount)
	assert.Equal(t, "SS-1", file.Rows[0].Values[ColumnAssetID])
	assert.Equal(t, "15/2/09", file.Rows[0].Values[ColumnCommissionedDate])
}

func TestParse_HeaderByNameInAnyOrderAndExtraColumnsIgnored(t *testing.T) {
	body := "asset_name,extra,commissioned_date,serial_number,model,manufacturer,rating_kva,voltage_kv," +
		"operational_status,asset_type,parent_asset_id,asset_id\n" +
		"Sub One,ignored,15/2/09,S-1,,,,22,IN_SERVICE,SUBSTATION,,SS-1\n"

	file, fileErr := Parse(strings.NewReader(body), testMaxBytes)

	require.Nil(t, fileErr)
	assert.Equal(t, 12, file.HeaderFieldCount)
	row := file.Rows[0]
	assert.Equal(t, "SS-1", row.Values[ColumnAssetID])
	assert.Equal(t, "Sub One", row.Values[ColumnAssetName])
	assert.Equal(t, "SUBSTATION", row.Values[ColumnAssetType])
	assert.NotContains(t, row.Values, Column("extra"))
}

func TestParse_HeaderNamesAreTrimmed(t *testing.T) {
	body := strings.ReplaceAll(testHeader, ",", " , ") + "\n" + substationLine + "\n"

	file, fileErr := Parse(strings.NewReader(body), testMaxBytes)

	require.Nil(t, fileErr)
	assert.Equal(t, "SS-1", file.Rows[0].Values[ColumnAssetID])
}

func TestParse_BOMIsAccepted(t *testing.T) {
	bom := string([]byte{0xEF, 0xBB, 0xBF})
	file, fileErr := Parse(strings.NewReader(bom+csvText(substationLine)), testMaxBytes)

	require.Nil(t, fileErr)
	assert.Equal(t, "SS-1", file.Rows[0].Values[ColumnAssetID])
}

func TestParse_QuotedFields(t *testing.T) {
	quoted := "SS-1,,SUBSTATION,\"Sub, \"\"North\"\"\nwing\",IN_SERVICE,22,,,,S-1,15/2/09"
	next := "SS-2,,SUBSTATION,Sub Two,IN_SERVICE,22,,,,S-2,15/2/09"

	file, fileErr := Parse(strings.NewReader(csvText(quoted, next)), testMaxBytes)

	require.Nil(t, fileErr)
	require.Len(t, file.Rows, 2)
	assert.Equal(t, "Sub, \"North\"\nwing", file.Rows[0].Values[ColumnAssetName])
	assert.Equal(t, 2, file.Rows[0].Number, "a record spanning lines is numbered by its first line")
	assert.Equal(t, 4, file.Rows[1].Number)
}

func TestParse_WrongFieldCountIsKeptAsARow(t *testing.T) {
	file, fileErr := Parse(strings.NewReader(csvText("SS-1,,SUBSTATION", substationLine)), testMaxBytes)

	require.Nil(t, fileErr)
	require.Len(t, file.Rows, 2)
	assert.Equal(t, 3, file.Rows[0].FieldCount)
	assert.Equal(t, "SUBSTATION", file.Rows[0].Values[ColumnAssetType])
	assert.Equal(t, "", file.Rows[0].Values[ColumnAssetName])
	assert.Equal(t, 11, file.Rows[1].FieldCount)
}

func TestParse_BlankLinesAreSkippedButCounted(t *testing.T) {
	body := testHeader + "\n\n" + substationLine + "\n"

	file, fileErr := Parse(strings.NewReader(body), testMaxBytes)

	require.Nil(t, fileErr)
	require.Len(t, file.Rows, 1)
	assert.Equal(t, 3, file.Rows[0].Number)
}

func TestParse_FileErrors(t *testing.T) {
	missingTwo := strings.Replace(strings.Replace(testHeader, ",asset_name", "", 1), ",model", "", 1)

	tests := []struct {
		name     string
		body     string
		maxBytes int64
		want     FileError
	}{
		{"empty file", "", testMaxBytes, FileError{Code: errorcode.ImportFileEmpty}},
		{"whitespace only", " \n\n", testMaxBytes, FileError{Code: errorcode.ImportFileEmpty}},
		{"header only", testHeader + "\n", testMaxBytes, FileError{Code: errorcode.ImportFileEmpty}},
		{
			"missing columns in CSV order", missingTwo + "\n" + substationLine + "\n", testMaxBytes,
			FileError{Code: errorcode.ImportMissingColumns, Args: []any{"asset_name, model"}},
		},
		{
			"repeated column", testHeader + ",asset_id\n" + substationLine + ",SS-1\n", testMaxBytes,
			FileError{Code: errorcode.ImportDuplicateColumns, Args: []any{"asset_id"}},
		},
		{
			"invalid UTF-8", csvText(substationLine, "SS-2,,SUBSTATION,Bad \xff name,,,,,,,"), testMaxBytes,
			FileError{Code: errorcode.ImportFileNotCSV, Args: []any{3}},
		},
		{
			"unclosed quote", csvText(substationLine, "SS-2,,SUBSTATION,\"Never closed,,,,,,,"), testMaxBytes,
			FileError{Code: errorcode.ImportFileNotCSV, Args: []any{3}},
		},
		{
			"stray quote in unquoted field", csvText(substationLine, substationLine, "SS-3,,SUBSTATION,Sub \"3\",,,,,,,"), testMaxBytes,
			FileError{Code: errorcode.ImportFileNotCSV, Args: []any{4}},
		},
		{"too large", csvText(substationLine), 10, FileError{Code: errorcode.ImportFileTooLarge, Args: []any{"10 bytes"}}},
		{
			"too large in MB", strings.Repeat("x", 2*1024*1024+1), 2 * 1024 * 1024,
			FileError{Code: errorcode.ImportFileTooLarge, Args: []any{"2 MB"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, fileErr := Parse(strings.NewReader(tt.body), tt.maxBytes)

			require.NotNil(t, fileErr)
			assert.Equal(t, tt.want, *fileErr)
			assert.Empty(t, file.Rows, "a refused file returns no rows")
		})
	}
}

func TestParse_ExactlyMaxBytesIsAccepted(t *testing.T) {
	body := csvText(substationLine)

	_, fileErr := Parse(strings.NewReader(body), int64(len(body)))

	assert.Nil(t, fileErr)
}
